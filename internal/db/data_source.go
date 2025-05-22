package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db/repository"
	"log"
)

type DataSource struct {
	pool    *pgxpool.Pool
	Queries *repository.Queries
}

func NewDataSource(dbConfig config.DbConfig) *DataSource {
	connString := fmt.Sprintf("postgres://%s:%s@%s",
		dbConfig.DBUser,
		dbConfig.DBPassword,
		dbConfig.DBUrl)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatal("Unable to parse config: ", err)
	}

	poolConfig.MaxConns = int32(dbConfig.DBMaxConns)
	poolConfig.MinConns = int32(dbConfig.DBMinConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Unable to create connection pool: ", err)
	}

	var result string
	err = pool.QueryRow(context.Background(), "select 'db connection initialized'").Scan(&result)
	if err != nil {
		pool.Close()
		log.Fatal("Check query failed:", err)
	}

	log.Println(result)
	return &DataSource{pool, repository.New(pool)}
}

func (ds *DataSource) Close() {
	ds.pool.Close()
}

func (ds *DataSource) ExecTx(ctx context.Context, fn func(*repository.Queries) error) error {
	tx, err := ds.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}

	q := ds.Queries.WithTx(tx)

	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}

	return tx.Commit(ctx)
}

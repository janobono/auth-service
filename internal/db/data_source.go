package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/janobono/auth-service/gen/db/repository"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db/dal"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DataSource struct {
	pool       *pgxpool.Pool
	Queries    *repository.Queries
	DalQueries *dal.Queries
}

func NewDataSource(dbConfig *config.DbConfig) *DataSource {
	slog.Info("Connecting to database", "url", dbConfig.Url)

	connString := fmt.Sprintf("postgres://%s:%s@%s", dbConfig.User, dbConfig.Password, dbConfig.Url)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		slog.Error("Unable to parse config", "error", err)
		panic(err)
	}

	poolConfig.MaxConns = int32(dbConfig.MaxConnections)
	poolConfig.MinConns = int32(dbConfig.MinConnections)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		slog.Error("Unable to create connection pool", "error", err)
		panic(err)
	}

	var result string
	err = pool.QueryRow(context.Background(), "select 'Database connection initialized'").Scan(&result)
	if err != nil {
		pool.Close()
		slog.Error("Unable to check connection", "error", err)
		panic(err)
	}

	slog.Info(result)

	stdlib.RegisterConnConfig(poolConfig.ConnConfig)
	db := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error("Unable to initialize migration driver", "error", err)
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./db/migrations", "postgres", driver)
	if err != nil {
		slog.Error("Unable to initialize migrations", "error", err)
		panic(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("Unable to execute migrations", "error", err)
		panic(err)
	}

	slog.Info("Migrations applied")

	return &DataSource{pool, repository.New(pool), dal.New(pool)}
}

func (ds *DataSource) Close() {
	ds.pool.Close()
}

func (ds *DataSource) ExecTx(ctx context.Context, fn func(*repository.Queries) (interface{}, error)) (interface{}, error) {
	tx, err := ds.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, err
	}

	q := ds.Queries.WithTx(tx)

	result, err := fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return nil, fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return nil, err
	}

	return result, tx.Commit(ctx)
}

func (ds *DataSource) ExecTxDal(ctx context.Context, fn func(*dal.Queries) (interface{}, error)) (interface{}, error) {
	tx, err := ds.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, err
	}

	q := ds.DalQueries.WithTx(tx)

	result, err := fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return nil, fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return nil, err
	}

	return result, tx.Commit(ctx)
}

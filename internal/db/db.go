package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/janobono/auth-service/internal/config"
	"log"
)

func InitDb(dbConfig config.DbConfig) *pgxpool.Pool {
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
	return pool
}

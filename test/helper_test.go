package test

import (
	"context"
	"fmt"
	"github.com/janobono/auth-service/internal/config"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	TestConfig *config.ServerConfig
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgres, cfg, err := startPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}

	TestConfig = &config.ServerConfig{
		DbConfig: *cfg,
	}

	code := m.Run()

	_ = postgres.Terminate(ctx)

	os.Exit(code)
}

func startPostgresContainer(ctx context.Context) (tc.Container, *config.DbConfig, error) {
	req := tc.ContainerRequest{
		Image:        "postgres:alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "app",
			"POSTGRES_USER":     "app",
			"POSTGRES_DB":       "app",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "pgx", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=app password=app dbname=app sslmode=disable", host, port.Port())
		}).WithStartupTimeout(30 * time.Second),
	}

	postgres, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	host, err := postgres.Host(ctx)
	if err != nil {
		return nil, nil, err
	}
	p, err := postgres.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}

	return postgres, &config.DbConfig{
		Url:            fmt.Sprintf("%s:%s/app", host, p.Port()),
		User:           "app",
		Password:       "app",
		MaxConnections: 5,
		MinConnections: 2,
	}, nil
}

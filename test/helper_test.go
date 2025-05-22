package test

import (
	"context"
	"fmt"
	"github.com/janobono/auth-service/internal/config"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	TestConfig *config.Config
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgres, cfg, err := startPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}

	port, err := getFreePort()
	if err != nil {
		log.Fatalf("could not find free port: %v", err)
	}

	TestConfig = &config.Config{
		ServerConfig: config.ServerConfig{
			Address: port,
		},
		DbConfig: *cfg,
		AppConfig: config.AppConfig{
			TokenExpiresIn: 30,
			TokenIssuer:    "test",
		},
	}

	code := m.Run()

	_ = postgres.Terminate(ctx)

	os.Exit(code)
}

func startPostgresContainer(ctx context.Context) (tc.Container, *config.DbConfig, error) {
	absPath, err := filepath.Abs("../db/init.sql")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	req := tc.ContainerRequest{
		Image:        "postgres:alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "app",
			"POSTGRES_USER":     "app",
			"POSTGRES_DB":       "app",
		},
		Files: []tc.ContainerFile{
			{
				HostFilePath:      absPath,
				ContainerFilePath: "/docker-entrypoint-initdb.d/init.sql",
				FileMode:          0644,
			},
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
		DBUrl:      fmt.Sprintf("%s:%s/app", host, p.Port()),
		DBUser:     "app",
		DBPassword: "app",
		DBMaxConns: 5,
		DBMinConns: 2,
	}, nil
}

func getFreePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	addr := l.Addr().(*net.TCPAddr)
	return fmt.Sprintf(":%d", addr.Port), nil
}

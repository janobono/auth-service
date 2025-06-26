package server

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/janobono/auth-service/generated/sqlc"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"log/slog"

	janobonodb "github.com/janobono/go-util/db"
)

func initDefaultCredentials(config *config.ServerConfig, dataSource *db.DataSource) {
	slog.Info("Initializing default credentials")

	defaultAuthorities := initDefaultAuthorities(dataSource, append(
		config.SecurityConfig.ReadAuthorities,
		config.SecurityConfig.WriteAuthorities...,
	))

	count, err := dataSource.Queries.CountAllUsers(context.Background())
	if err != nil {
		slog.Error("Failed to count users", "error", err)
		panic(err)
	}
	if count > 0 {
		slog.Info("Default credentials already initialized")
		return
	}

	_, err = dataSource.Queries.GetUserByEmail(context.Background(), config.SecurityConfig.DefaultUsername)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.Error("Failed to check default user", "error", err)
		panic(err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		newUser, err := dataSource.Queries.AddUser(context.Background(), sqlc.AddUserParams{
			ID:        janobonodb.NewUUID(),
			CreatedAt: janobonodb.NowUTC(),
			Email:     config.SecurityConfig.DefaultUsername,
			Password:  config.SecurityConfig.DefaultPassword,
			Confirmed: true,
			Enabled:   true,
		})

		if err != nil {
			slog.Error("Failed to create default user", "error", err)
			panic(err)
		}

		slog.Info("Default user created", "email", config.SecurityConfig.DefaultUsername)

		for _, defaultAuthority := range defaultAuthorities {
			err := dataSource.Queries.AddUserAuthority(context.Background(), sqlc.AddUserAuthorityParams{
				UserID:      newUser.ID,
				AuthorityID: defaultAuthority.ID,
			})

			if err != nil {
				slog.Error("Failed to add user authority", "error", err)
				panic(err)
			}
		}
	}

	slog.Info("Default credentials initialized")
}

func initDefaultAuthorities(dataSource *db.DataSource, defaultAuthorities []string) []sqlc.Authority {
	var result []sqlc.Authority
	slog.Info("Initializing default authorities")

	for _, authority := range defaultAuthorities {
		savedAuthority, err := dataSource.Queries.GetAuthority(context.Background(), authority)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Failed to get authority", "error", err)
			panic(err)
		}

		if errors.Is(err, pgx.ErrNoRows) {
			newAuthority, err := dataSource.Queries.AddAuthority(context.Background(), sqlc.AddAuthorityParams{
				ID:        janobonodb.NewUUID(),
				Authority: authority,
			})
			if err != nil {
				slog.Error("Failed to add authority", "error", err)
				panic(err)
			}

			slog.Info("Added authority", "authority", authority)
			result = append(result, newAuthority)
			continue
		}

		result = append(result, savedAuthority)
	}

	slog.Info("Default authorities initialized")
	return result
}

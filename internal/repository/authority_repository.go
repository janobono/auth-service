package repository

import (
	"context"
	"github.com/janobono/auth-service/generated/sqlc"
	"github.com/janobono/auth-service/internal/db"
	db2 "github.com/janobono/go-util/db"
)

type AuthorityRepository interface {
	AddAuthority(ctx context.Context, arg AddAuthorityData) (*Authority, error)
	DeleteAuthority(ctx context.Context, id string) error
	GetAuthority(ctx context.Context, authority string) (*Authority, error)
}

type authorityRepositoryImpl struct {
	dataSource *db.DataSource
}

func NewAuthorityRepository(dataSource *db.DataSource) AuthorityRepository {
	return &authorityRepositoryImpl{dataSource}
}

func (u *authorityRepositoryImpl) AddAuthority(ctx context.Context, arg AddAuthorityData) (*Authority, error) {
	authority, err := u.dataSource.Queries.AddAuthority(ctx, sqlc.AddAuthorityParams{
		ID:        db2.NewUUID(),
		Authority: arg.Authority,
	})

	if err != nil {
		return nil, err
	}

	return toAuthority(&authority), nil
}

func (u *authorityRepositoryImpl) DeleteAuthority(ctx context.Context, id string) error {
	pgId, err := db2.ParseUUID(id)

	if err != nil {
		return err
	}

	return u.dataSource.Queries.DeleteAuthority(ctx, pgId)
}

func (u *authorityRepositoryImpl) GetAuthority(ctx context.Context, authority string) (*Authority, error) {
	dbAuthority, err := u.dataSource.Queries.GetAuthority(ctx, authority)

	if err != nil {
		return nil, err
	}

	return toAuthority(&dbAuthority), nil
}

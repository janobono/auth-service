package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/sqlc"
	"github.com/janobono/auth-service/internal/db"
	db2 "github.com/janobono/go-util/db"
)

type AuthorityRepository interface {
	AddAuthority(ctx context.Context, arg AuthorityData) (*Authority, error)
	DeleteAuthority(ctx context.Context, id pgtype.UUID) error
	GetAuthority(ctx context.Context, authority string) (*Authority, error)
}

type authorityRepositoryImpl struct {
	dataSource *db.DataSource
}

func NewAuthorityRepository(dataSource *db.DataSource) AuthorityRepository {
	return &authorityRepositoryImpl{dataSource}
}

func (u *authorityRepositoryImpl) AddAuthority(ctx context.Context, arg AuthorityData) (*Authority, error) {
	authority, err := u.dataSource.Queries.AddAuthority(ctx, sqlc.AddAuthorityParams{
		ID:        db2.NewUUID(),
		Authority: arg.Authority,
	})

	if err != nil {
		return nil, err
	}

	return toAuthority(&authority), nil
}

func (u *authorityRepositoryImpl) DeleteAuthority(ctx context.Context, id pgtype.UUID) error {
	return u.dataSource.Queries.DeleteAuthority(ctx, id)
}

func (u *authorityRepositoryImpl) GetAuthority(ctx context.Context, authority string) (*Authority, error) {
	dbAuthority, err := u.dataSource.Queries.GetAuthority(ctx, authority)

	if err != nil {
		return nil, err
	}

	return toAuthority(&dbAuthority), nil
}

package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/sqlc"
	"github.com/janobono/auth-service/internal/db"
	db2 "github.com/janobono/go-util/db"
)

type AttributeRepository interface {
	AddAttribute(ctx context.Context, arg AddAttributeData) (*Attribute, error)
	DeleteAttribute(ctx context.Context, id pgtype.UUID) error
	GetAttribute(ctx context.Context, key string) (*Attribute, error)
}

type attributeRepositoryImpl struct {
	dataSource *db.DataSource
}

func NewAttributeRepository(dataSource *db.DataSource) AttributeRepository {
	return &attributeRepositoryImpl{dataSource}
}

func (a *attributeRepositoryImpl) AddAttribute(ctx context.Context, arg AddAttributeData) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.AddAttribute(ctx, sqlc.AddAttributeParams{
		ID:       db2.NewUUID(),
		Key:      arg.Key,
		Required: arg.Required,
		Hidden:   arg.Hidden,
	})

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

func (a *attributeRepositoryImpl) DeleteAttribute(ctx context.Context, id pgtype.UUID) error {
	return a.dataSource.Queries.DeleteAttribute(ctx, id)
}

func (a *attributeRepositoryImpl) GetAttribute(ctx context.Context, key string) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.GetAttribute(ctx, key)

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

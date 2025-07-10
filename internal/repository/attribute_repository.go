package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/sqlc"
	"github.com/janobono/auth-service/internal/db"
	db2 "github.com/janobono/go-util/db"
)

type AttributeRepository interface {
	AddAttribute(ctx context.Context, attributeData AttributeData) (*Attribute, error)
	CountById(ctx context.Context, id pgtype.UUID) (int64, error)
	CountByKey(ctx context.Context, key string) (int64, error)
	DeleteAttribute(ctx context.Context, id pgtype.UUID) error
	GetAttribute(ctx context.Context, key string) (*Attribute, error)
	SetAttribute(ctx context.Context, id pgtype.UUID, attributeData AttributeData) (*Attribute, error)
}

type attributeRepositoryImpl struct {
	dataSource *db.DataSource
}

func NewAttributeRepository(dataSource *db.DataSource) AttributeRepository {
	return &attributeRepositoryImpl{dataSource}
}

func (a *attributeRepositoryImpl) AddAttribute(ctx context.Context, attributeData AttributeData) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.AddAttribute(ctx, sqlc.AddAttributeParams{
		ID:       db2.NewUUID(),
		Key:      attributeData.Key,
		Required: attributeData.Required,
		Hidden:   attributeData.Hidden,
	})

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

func (a *attributeRepositoryImpl) CountById(ctx context.Context, id pgtype.UUID) (int64, error) {
	return a.dataSource.Queries.CountAttributeById(ctx, id)
}

func (a *attributeRepositoryImpl) CountByKey(ctx context.Context, key string) (int64, error) {
	return a.dataSource.Queries.CountAttributeByKey(ctx, key)
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

func (a *attributeRepositoryImpl) SetAttribute(ctx context.Context, id pgtype.UUID, attributeData AttributeData) (*Attribute, error) {
	attribute, err := a.dataSource.Queries.SetAttribute(ctx, sqlc.SetAttributeParams{
		ID:       id,
		Key:      attributeData.Key,
		Required: attributeData.Required,
		Hidden:   attributeData.Hidden,
	})

	if err != nil {
		return nil, err
	}

	return toAttribute(&attribute), nil
}

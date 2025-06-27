package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
)

type AttributeService interface {
	AddAttribute(ctx context.Context, attributeData *openapi.AttributeData) (*openapi.AttributeDetail, error)
	DeleteAttribute(ctx context.Context, id pgtype.UUID) error
	GetAttribute(ctx context.Context, id pgtype.UUID) (*openapi.AttributeDetail, error)
	GetAttributes(ctx context.Context, criteria *SearchAttributeCriteria, pageable *common.Pageable) (*common.Page[*openapi.AttributeDetail], error)
	SetAttribute(ctx context.Context, id pgtype.UUID, data *openapi.AttributeData) (*openapi.AttributeDetail, error)
}

type attributeService struct {
	attributeRepository repository.AttributeRepository
}

var _ AttributeService = (*attributeService)(nil)

func NewAttributeService(attributeRepository repository.AttributeRepository) AttributeService {
	return &attributeService{attributeRepository}
}

func (a attributeService) AddAttribute(ctx context.Context, attributeData *openapi.AttributeData) (*openapi.AttributeDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (a attributeService) DeleteAttribute(ctx context.Context, id pgtype.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (a attributeService) GetAttribute(ctx context.Context, id pgtype.UUID) (*openapi.AttributeDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (a attributeService) GetAttributes(ctx context.Context, criteria *SearchAttributeCriteria, pageable *common.Pageable) (*common.Page[*openapi.AttributeDetail], error) {
	//TODO implement me
	panic("implement me")
}

func (a attributeService) SetAttribute(ctx context.Context, id pgtype.UUID, data *openapi.AttributeData) (*openapi.AttributeDetail, error) {
	//TODO implement me
	panic("implement me")
}

package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"net/http"
)

type AttributeService interface {
	AddAttribute(ctx context.Context, data *openapi.AttributeData) (*openapi.AttributeDetail, error)
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

func (a *attributeService) AddAttribute(ctx context.Context, data *openapi.AttributeData) (*openapi.AttributeDetail, error) {
	count, err := a.attributeRepository.CountByKey(ctx, data.Key)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'key' already exists")
	}

	attribute, err := a.attributeRepository.AddAttribute(ctx, &repository.AttributeData{
		Key:      data.Key,
		Required: data.Required,
		Hidden:   data.Hidden,
	})
	if err != nil {
		return nil, err
	}

	return &openapi.AttributeDetail{
		Id:       attribute.ID.String(),
		Key:      attribute.Key,
		Required: attribute.Required,
		Hidden:   attribute.Hidden,
	}, nil
}

func (a *attributeService) DeleteAttribute(ctx context.Context, id pgtype.UUID) error {
	count, err := a.attributeRepository.CountById(ctx, id)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "attribute does not exist")
	}

	return a.attributeRepository.DeleteAttributeById(ctx, id)
}

func (a *attributeService) GetAttribute(ctx context.Context, id pgtype.UUID) (*openapi.AttributeDetail, error) {
	attribute, err := a.attributeRepository.GetAttributeById(ctx, id)
	if err != nil {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "attribute not found")
	}

	return &openapi.AttributeDetail{
		Id:       attribute.ID.String(),
		Key:      attribute.Key,
		Required: attribute.Required,
		Hidden:   attribute.Hidden,
	}, nil
}

func (a *attributeService) GetAttributes(ctx context.Context, criteria *SearchAttributeCriteria, pageable *common.Pageable) (*common.Page[*openapi.AttributeDetail], error) {
	page, err := a.attributeRepository.SearchAttributes(ctx, &repository.SearchAttributesCriteria{
		SearchField: criteria.SearchField,
	}, pageable)
	if err != nil {
		return nil, err
	}

	content := make([]*openapi.AttributeDetail, len(page.Content))
	for i, attribute := range page.Content {
		content[i] = &openapi.AttributeDetail{
			Id:       attribute.ID.String(),
			Key:      attribute.Key,
			Required: attribute.Required,
			Hidden:   attribute.Hidden,
		}
	}

	return &common.Page[*openapi.AttributeDetail]{
		Pageable:      pageable,
		TotalElements: page.TotalElements,
		TotalPages:    page.TotalPages,
		First:         page.First,
		Last:          page.Last,
		Content:       content,
		Empty:         page.Empty,
	}, nil
}

func (a *attributeService) SetAttribute(ctx context.Context, id pgtype.UUID, data *openapi.AttributeData) (*openapi.AttributeDetail, error) {
	count, err := a.attributeRepository.CountById(ctx, id)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "attribute does not exist")
	}

	count, err = a.attributeRepository.CountByKeyAndNotId(ctx, data.Key, id)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'key' already exists")
	}

	attribute, err := a.attributeRepository.SetAttribute(ctx, id, &repository.AttributeData{
		Key:      data.Key,
		Required: data.Required,
		Hidden:   data.Hidden,
	})
	if err != nil {
		return nil, err
	}

	return &openapi.AttributeDetail{
		Id:       attribute.ID.String(),
		Key:      attribute.Key,
		Required: attribute.Required,
		Hidden:   attribute.Hidden,
	}, nil
}

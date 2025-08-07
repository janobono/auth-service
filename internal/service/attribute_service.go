package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
)

type AttributeService struct {
	attributeRepository repository.AttributeRepository
}

func NewAttributeService(attributeRepository repository.AttributeRepository) *AttributeService {
	return &AttributeService{attributeRepository}
}

func (as *AttributeService) AddAttribute(ctx context.Context, data *openapi.AttributeData) (*openapi.AttributeDetail, error) {
	count, err := as.attributeRepository.CountByKey(ctx, data.Key)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'key' already exists")
	}

	attribute, err := as.attributeRepository.AddAttribute(ctx, &repository.AttributeData{
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

func (as *AttributeService) DeleteAttribute(ctx context.Context, id pgtype.UUID) error {
	count, err := as.attributeRepository.CountById(ctx, id)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "attribute does not exist")
	}

	return as.attributeRepository.DeleteAttributeById(ctx, id)
}

func (as *AttributeService) GetAttribute(ctx context.Context, id pgtype.UUID) (*openapi.AttributeDetail, error) {
	attribute, err := as.attributeRepository.GetAttributeById(ctx, id)
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

func (as *AttributeService) GetAttributes(ctx context.Context, criteria *SearchAttributeCriteria, pageable *common.Pageable) (*common.Page[*openapi.AttributeDetail], error) {
	page, err := as.attributeRepository.SearchAttributes(ctx, &repository.SearchAttributesCriteria{
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

func (as *AttributeService) SetAttribute(ctx context.Context, id pgtype.UUID, data *openapi.AttributeData) (*openapi.AttributeDetail, error) {
	count, err := as.attributeRepository.CountById(ctx, id)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "attribute does not exist")
	}

	count, err = as.attributeRepository.CountByKeyAndNotId(ctx, data.Key, id)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'key' already exists")
	}

	attribute, err := as.attributeRepository.SetAttribute(ctx, id, &repository.AttributeData{
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

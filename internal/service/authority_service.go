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

type AuthorityService interface {
	AddAuthority(ctx context.Context, data *openapi.AuthorityData) (*openapi.AuthorityDetail, error)
	DeleteAuthority(ctx context.Context, id pgtype.UUID) error
	GetAuthority(ctx context.Context, id pgtype.UUID) (*openapi.AuthorityDetail, error)
	GetAuthorities(ctx context.Context, criteria *SearchAuthorityCriteria, pageable *common.Pageable) (*common.Page[*openapi.AuthorityDetail], error)
	SetAuthority(ctx context.Context, id pgtype.UUID, data *openapi.AuthorityData) (*openapi.AuthorityDetail, error)
}

type authorityService struct {
	authorityRepository repository.AuthorityRepository
}

var _ AuthorityService = (*authorityService)(nil)

func NewAuthorityService(authorityRepository repository.AuthorityRepository) AuthorityService {
	return &authorityService{authorityRepository}
}

func (a *authorityService) AddAuthority(ctx context.Context, data *openapi.AuthorityData) (*openapi.AuthorityDetail, error) {
	count, err := a.authorityRepository.CountByAuthority(ctx, data.Authority)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'authority' already exists")
	}

	authority, err := a.authorityRepository.AddAuthority(ctx, &repository.AuthorityData{
		Authority: data.Authority,
	})
	if err != nil {
		return nil, err
	}

	return &openapi.AuthorityDetail{
		Id:        authority.ID.String(),
		Authority: authority.Authority,
	}, nil
}

func (a *authorityService) DeleteAuthority(ctx context.Context, id pgtype.UUID) error {
	count, err := a.authorityRepository.CountById(ctx, id)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "authority does not exist")
	}

	return a.authorityRepository.DeleteAuthorityById(ctx, id)
}

func (a *authorityService) GetAuthority(ctx context.Context, id pgtype.UUID) (*openapi.AuthorityDetail, error) {
	authority, err := a.authorityRepository.GetAuthorityById(ctx, id)
	if err != nil {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "authority not found")
	}

	return &openapi.AuthorityDetail{
		Id:        authority.ID.String(),
		Authority: authority.Authority,
	}, nil
}

func (a *authorityService) GetAuthorities(ctx context.Context, criteria *SearchAuthorityCriteria, pageable *common.Pageable) (*common.Page[*openapi.AuthorityDetail], error) {
	page, err := a.authorityRepository.SearchAuthorities(ctx, &repository.SearchAuthoritiesCriteria{
		SearchField: criteria.SearchField,
	}, pageable)
	if err != nil {
		return nil, err
	}

	content := make([]*openapi.AuthorityDetail, len(page.Content))
	for i, authority := range page.Content {
		content[i] = &openapi.AuthorityDetail{
			Id:        authority.ID.String(),
			Authority: authority.Authority,
		}
	}

	return &common.Page[*openapi.AuthorityDetail]{
		Pageable:      pageable,
		TotalElements: page.TotalElements,
		TotalPages:    page.TotalPages,
		First:         page.First,
		Last:          page.Last,
		Content:       content,
		Empty:         page.Empty,
	}, nil
}

func (a *authorityService) SetAuthority(ctx context.Context, id pgtype.UUID, data *openapi.AuthorityData) (*openapi.AuthorityDetail, error) {
	count, err := a.authorityRepository.CountById(ctx, id)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "authority does not exist")
	}

	count, err = a.authorityRepository.CountByAuthorityAndNotId(ctx, data.Authority, id)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'authority' already exists")
	}

	authority, err := a.authorityRepository.SetAuthority(ctx, id, &repository.AuthorityData{
		Authority: data.Authority,
	})
	if err != nil {
		return nil, err
	}

	return &openapi.AuthorityDetail{
		Id:        authority.ID.String(),
		Authority: authority.Authority,
	}, nil
}

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

type AuthorityService struct {
	authorityRepository repository.AuthorityRepository
}

func NewAuthorityService(authorityRepository repository.AuthorityRepository) *AuthorityService {
	return &AuthorityService{authorityRepository}
}

func (as *AuthorityService) AddAuthority(ctx context.Context, data *openapi.AuthorityData) (*openapi.AuthorityDetail, error) {
	count, err := as.authorityRepository.CountByAuthority(ctx, data.Authority)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'authority' already exists")
	}

	authority, err := as.authorityRepository.AddAuthority(ctx, &repository.AuthorityData{
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

func (as *AuthorityService) DeleteAuthority(ctx context.Context, id pgtype.UUID) error {
	count, err := as.authorityRepository.CountById(ctx, id)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "authority does not exist")
	}

	return as.authorityRepository.DeleteAuthorityById(ctx, id)
}

func (as *AuthorityService) GetAuthority(ctx context.Context, id pgtype.UUID) (*openapi.AuthorityDetail, error) {
	authority, err := as.authorityRepository.GetAuthorityById(ctx, id)
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

func (as *AuthorityService) GetAuthorities(ctx context.Context, criteria *SearchAuthorityCriteria, pageable *common.Pageable) (*common.Page[*openapi.AuthorityDetail], error) {
	page, err := as.authorityRepository.SearchAuthorities(ctx, &repository.SearchAuthoritiesCriteria{
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

func (as *AuthorityService) SetAuthority(ctx context.Context, id pgtype.UUID, data *openapi.AuthorityData) (*openapi.AuthorityDetail, error) {
	count, err := as.authorityRepository.CountById(ctx, id)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "authority does not exist")
	}

	count, err = as.authorityRepository.CountByAuthorityAndNotId(ctx, data.Authority, id)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'authority' already exists")
	}

	authority, err := as.authorityRepository.SetAuthority(ctx, id, &repository.AuthorityData{
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

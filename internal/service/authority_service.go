package service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
)

type AuthorityService interface {
	AddAuthority(ctx context.Context, AuthorityData *openapi.AuthorityData) (*openapi.AuthorityDetail, error)
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

func (a *authorityService) AddAuthority(ctx context.Context, AuthorityData *openapi.AuthorityData) (*openapi.AuthorityDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (a *authorityService) DeleteAuthority(ctx context.Context, id pgtype.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (a *authorityService) GetAuthority(ctx context.Context, id pgtype.UUID) (*openapi.AuthorityDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (a *authorityService) GetAuthorities(ctx context.Context, criteria *SearchAuthorityCriteria, pageable *common.Pageable) (*common.Page[*openapi.AuthorityDetail], error) {
	//TODO implement me
	panic("implement me")
}

func (a *authorityService) SetAuthority(ctx context.Context, id pgtype.UUID, data *openapi.AuthorityData) (*openapi.AuthorityDetail, error) {
	//TODO implement me
	panic("implement me")
}

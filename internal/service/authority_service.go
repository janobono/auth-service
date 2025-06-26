package service

import "github.com/janobono/auth-service/internal/repository"

type AuthorityService interface {
}

type authorityService struct {
	authorityRepository repository.AuthorityRepository
}

var _ AuthorityService = (*authorityService)(nil)

func NewAuthorityService(authorityRepository repository.AuthorityRepository) AuthorityService {
	return &authorityService{authorityRepository}
}

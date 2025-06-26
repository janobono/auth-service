package service

import (
	"context"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
)

type JwkService interface {
	GetActiveJwks(ctx context.Context) ([]*Jwk, error)
}

type jwkService struct {
	jwkRepository repository.JwkRepository
}

var _ JwkService = (*jwkService)(nil)

func NewJwkService(jwkRepository repository.JwkRepository) JwkService {
	return &jwkService{jwkRepository}
}

func (j *jwkService) GetActiveJwks(ctx context.Context) ([]*Jwk, error) {
	activeJwks, err := j.jwkRepository.GetActiveJwks(ctx)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	result := make([]*Jwk, 0, len(activeJwks))

	for i, dbJwk := range activeJwks {
		result[i] = &Jwk{
			Id:         dbJwk.ID,
			Kty:        dbJwk.Kty,
			Use:        dbJwk.Use,
			Alg:        dbJwk.Alg,
			PublicKey:  dbJwk.PublicKey,
			PrivateKey: dbJwk.PrivateKey,
			Active:     dbJwk.Active,
			CreatedAt:  dbJwk.CreatedAt,
			ExpiresAt:  dbJwk.ExpiresAt,
		}
	}

	return result, nil
}

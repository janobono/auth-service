package service

import (
	"context"
	"encoding/base64"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"math/big"
)

type JwkService interface {
	GetJwks(ctx context.Context) (*openapi.Jwks, error)
}

type jwkService struct {
	jwkRepository repository.JwkRepository
}

var _ JwkService = (*jwkService)(nil)

func NewJwkService(jwkRepository repository.JwkRepository) JwkService {
	return &jwkService{jwkRepository}
}

func (j *jwkService) GetJwks(ctx context.Context) (*openapi.Jwks, error) {
	activeJwks, err := j.jwkRepository.GetActiveJwks(ctx)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	keys := make([]openapi.Jwk, 0, len(activeJwks))

	for i, jwk := range activeJwks {

		n := base64.RawURLEncoding.EncodeToString(jwk.PublicKey.N.Bytes())

		eBytes := big.NewInt(int64(jwk.PublicKey.E)).Bytes()
		if len(eBytes) < 4 {
			padding := make([]byte, 4-len(eBytes))
			eBytes = append(padding, eBytes...)
		}
		e := base64.RawURLEncoding.EncodeToString(eBytes)

		keys[i] = openapi.Jwk{
			Kty: jwk.Kty,
			Kid: jwk.ID.String(),
			Use: jwk.Use,
			Alg: jwk.Alg,
			N:   n,
			E:   e,
		}
	}

	return &openapi.Jwks{Keys: keys}, nil
}

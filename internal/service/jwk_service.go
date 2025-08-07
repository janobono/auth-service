package service

import (
	"context"
	"encoding/base64"
	"math/big"

	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
)

type JwkService struct {
	jwkRepository repository.JwkRepository
}

func NewJwkService(jwkRepository repository.JwkRepository) *JwkService {
	return &JwkService{jwkRepository}
}

func (js *JwkService) GetJwks(ctx context.Context) (*openapi.Jwks, error) {
	activeJwks, err := js.jwkRepository.GetActiveJwks(ctx)
	if err != nil {
		return nil, err
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

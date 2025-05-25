package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/janobono/auth-service/gen/db/repository"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/pkg/util"
	"time"
)

type JwkService struct {
	dataSource *db.DataSource
}

func NewJwkService(dataSource *db.DataSource) *JwkService {
	return &JwkService{dataSource}
}

func (j *JwkService) GetOrCreateJwk(ctx context.Context, use string, expiration int64) (*repository.Jwk, error) {
	jwk, err := j.dataSource.Queries.GetActiveJwk(ctx, use)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return j.CreateJwk(ctx, use, expiration)
	}
	return &jwk, nil
}

func (j *JwkService) CreateJwk(ctx context.Context, use string, expiration int64) (*repository.Jwk, error) {
	jwk, err := j.dataSource.ExecTx(ctx, func(q *repository.Queries) (interface{}, error) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		publicKey := &privateKey.PublicKey

		privatePEM := j.encodePrivateKey(privateKey)
		publicPEM, err := j.encodePublicKey(publicKey)
		if err != nil {
			return nil, err
		}

		now := time.Now()
		jwk, err := q.AddJwk(ctx, repository.AddJwkParams{
			ID:         util.NewUUID(),
			Kty:        "RSA",
			Use:        use,
			Alg:        "RS256",
			PublicKey:  publicPEM,
			PrivateKey: privatePEM,
			Active:     true,
			CreatedAt:  util.TimestampUTC(now),
			ExpiresAt:  util.TimestampUTC(now.Add(time.Duration(expiration) * time.Second)),
		})

		if err != nil {
			return nil, err
		}

		err = q.DeactivateJwks(ctx, repository.DeactivateJwksParams{ID: jwk.ID, Use: use})
		if err != nil {
			return nil, err
		}

		err = q.DeleteNotActiveJwks(ctx, repository.DeleteNotActiveJwksParams{
			Use:       use,
			ExpiresAt: util.NowUTC(),
		})
		if err != nil {
			return nil, err
		}

		return &jwk, nil
	})

	if err != nil {
		return nil, err
	}

	createdJwk, ok := jwk.(*repository.Jwk)
	if !ok {
		return nil, fmt.Errorf("invalid jwk type: %T", jwk)
	}

	return createdJwk, nil
}

func (j *JwkService) GetPublicKey(kid string) (interface{}, error) {
	uuid, err := util.ParseUUID(kid)
	if err != nil {
		return nil, err
	}

	jwk, err := j.dataSource.Queries.GetJwk(context.Background(), uuid)
	if err != nil {
		return nil, err
	}

	pubKey, err := j.parsePublicKey(&jwk)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

func (j *JwkService) CreateJwtToken(issuer string, tokenExpiration time.Duration, jwk *repository.Jwk) (*component.JwtToken, error) {
	privateKey, err := j.parsePrivate(jwk)
	if err != nil {
		return nil, err
	}

	publicKey, err := j.parsePublicKey(jwk)
	if err != nil {
		return nil, err
	}

	return component.NewJwtToken(
		jwt.SigningMethodRS256,
		privateKey,
		publicKey,
		jwk.ID.String(),
		issuer,
		tokenExpiration,
		jwk.ExpiresAt.Time,
		func(kid string) (interface{}, error) {
			return j.GetPublicKey(kid)
		},
	), nil
}

func (j *JwkService) encodePrivateKey(privateKey *rsa.PrivateKey) []byte {
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}
	return pem.EncodeToMemory(block)
}

func (j *JwkService) encodePublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	}
	return pem.EncodeToMemory(block), nil
}

func (j *JwkService) parsePrivate(jwk *repository.Jwk) (*rsa.PrivateKey, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(jwk.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	return privateKey, nil
}

func (j *JwkService) parsePublicKey(jwk *repository.Jwk) (*rsa.PublicKey, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(jwk.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}
	return publicKey, nil
}

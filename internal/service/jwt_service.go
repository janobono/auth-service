package service

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/repository"
	db2 "github.com/janobono/go-util/db"
	"github.com/janobono/go-util/security"
	"sync"
	"time"
)

type JwtService struct {
	securityConfig *config.SecurityConfig
	jwkRepository  repository.JwkRepository

	mutex        sync.Mutex
	accessToken  *security.JwtToken
	refreshToken *security.JwtToken
	confirmToken *security.JwtToken
}

func NewJwtService(securityConfig *config.SecurityConfig, jwkRepository repository.JwkRepository) *JwtService {
	return &JwtService{
		securityConfig: securityConfig,
		jwkRepository:  jwkRepository,
	}
}

func (j *JwtService) GetAccessJwtToken(ctx context.Context) (*security.JwtToken, error) {
	return j.getJwtToken(
		ctx,
		"access",
		j.securityConfig.AccessTokenExpiresIn,
		j.securityConfig.AccessTokenJwkExpiresIn,
		&j.accessToken,
	)
}

func (j *JwtService) GetRefreshJwtToken(ctx context.Context) (*security.JwtToken, error) {
	return j.getJwtToken(
		ctx,
		"refresh",
		j.securityConfig.RefreshTokenExpiresIn,
		j.securityConfig.RefreshTokenJwkExpiresIn,
		&j.refreshToken,
	)
}

func (j *JwtService) GetConfirmJwtToken(ctx context.Context) (*security.JwtToken, error) {
	return j.getJwtToken(
		ctx,
		"confirm",
		j.securityConfig.ContentTokenExpiresIn,
		j.securityConfig.ContentTokenJwkExpiresIn,
		&j.confirmToken,
	)
}

func (j *JwtService) getJwtToken(
	ctx context.Context,
	use string,
	tokenExpiration, jwkExpiration time.Duration,
	cached **security.JwtToken,
) (*security.JwtToken, error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	now := time.Now().UTC()

	if *cached != nil && now.Before((*cached).KeyExpiration()) {
		return *cached, nil
	}

	jwk, err := j.jwkRepository.GetActiveJwk(ctx, use)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if (err == nil && now.After(jwk.ExpiresAt)) || errors.Is(err, pgx.ErrNoRows) {
		jwk, err = j.jwkRepository.AddJwk(ctx, repository.JwkData{
			Use:        use,
			Expiration: jwkExpiration,
		})
	}

	if err != nil {
		return nil, err
	}

	token := security.NewJwtToken(
		jwt.SigningMethodRS256,
		jwk.PrivateKey,
		jwk.PublicKey,
		jwk.ID.String(),
		j.securityConfig.TokenIssuer,
		tokenExpiration,
		jwk.ExpiresAt,
		j.GetPublicKey,
	)

	*cached = token
	return token, nil
}

func (j *JwtService) GetPublicKey(ctx context.Context, kid string) (interface{}, error) {
	id, err := db2.ParseUUID(kid)
	if err != nil {
		return nil, err
	}

	jwk, err := j.jwkRepository.GetJwk(ctx, id)
	if err != nil {
		return nil, err
	}

	return jwk.PublicKey, nil
}

func (j *JwtService) GenerateAuthToken(token *security.JwtToken, id pgtype.UUID, authorities []string) (string, error) {
	claims := jwt.MapClaims{
		"sub": id.String(),
		"aud": authorities,
	}
	return token.GenerateToken(claims)
}

func (j *JwtService) ParseAuthToken(ctx context.Context, jwtToken *security.JwtToken, token string) (pgtype.UUID, []string, error) {
	claims, err := jwtToken.ParseToken(ctx, token)
	if err != nil {
		return pgtype.UUID{}, nil, err
	}

	idString, ok := (*claims)["sub"].(string)

	if !ok {
		return pgtype.UUID{}, nil, errors.New("invalid access token")
	}

	id, err := db2.ParseUUID(idString)
	if err != nil {
		return pgtype.UUID{}, nil, err
	}

	var authorities []string
	if aud, ok := (*claims)["aud"].([]interface{}); ok {
		for _, a := range aud {
			if aStr, ok := a.(string); ok {
				authorities = append(authorities, aStr)
			}
		}
	}

	return id, authorities, nil
}

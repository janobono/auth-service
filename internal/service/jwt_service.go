package service

import (
	"context"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"log/slog"
	"sync"
	"time"
)

type JwtService interface {
	GetAccessJwtToken() (*component.JwtToken, error)
	GetRefreshJwtToken() (*component.JwtToken, error)
	GetConfirmJwtToken() (*component.JwtToken, error)
}

type jwtService struct {
	securityConfig *config.SecurityConfig
	dataSource     *db.DataSource
	jwkService     *JwkService

	mutex        sync.Mutex
	accessToken  *component.JwtToken
	refreshToken *component.JwtToken
	confirmToken *component.JwtToken
}

func NewJwtService(securityConfig *config.SecurityConfig, dataSource *db.DataSource) JwtService {
	return &jwtService{
		securityConfig: securityConfig,
		dataSource:     dataSource,
		jwkService:     NewJwkService(dataSource),
	}
}

func (j *jwtService) GetAccessJwtToken() (*component.JwtToken, error) {
	return j.getJwtToken(
		"access",
		j.securityConfig.AccessTokenExpiresIn,
		j.securityConfig.AccessTokenJwkExpiresIn,
		&j.accessToken,
	)
}

func (j *jwtService) GetRefreshJwtToken() (*component.JwtToken, error) {
	return j.getJwtToken(
		"refresh",
		j.securityConfig.RefreshTokenExpiresIn,
		j.securityConfig.RefreshTokenJwkExpiresIn,
		&j.refreshToken,
	)
}

func (j *jwtService) GetConfirmJwtToken() (*component.JwtToken, error) {
	return j.getJwtToken(
		"confirm",
		j.securityConfig.ContentTokenExpiresIn,
		j.securityConfig.ContentTokenJwkExpiresIn,
		&j.confirmToken,
	)
}

func (j *jwtService) getJwtToken(
	use string,
	tokenExpiration, jwkExpiration time.Duration,
	cached **component.JwtToken,
) (*component.JwtToken, error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	now := time.Now().UTC()

	if *cached != nil && now.Before((*cached).KeyExpiration()) {
		return *cached, nil
	}

	ctx := context.Background()
	jwk, err := j.jwkService.GetOrCreateJwk(ctx, use, int64(jwkExpiration.Seconds()))
	if err != nil {
		slog.Error("Failed to get or create JWK", "use", use, "error", err)
		return nil, err
	}

	token, err := j.jwkService.CreateJwtToken(j.securityConfig.TokenIssuer, tokenExpiration, jwk)
	if err != nil {
		slog.Error("Failed to create JWT token", "use", use, "error", err)
		return nil, err
	}

	*cached = token
	return token, nil
}

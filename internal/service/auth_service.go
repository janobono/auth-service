package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/janobono/go-util/security"
)

type AuthService interface {
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	SignIn(ctx context.Context, signIn SignIn) (*AuthResponse, error)
}

type authService struct {
	passwordEncoder *security.PasswordEncoder
	jwtService      *JwtService
	userRepository  repository.UserRepository
}

var _ AuthService = (*authService)(nil)

func NewAuthService(
	passwordEncoder *security.PasswordEncoder,
	jwtService *JwtService,
	userRepository repository.UserRepository,
) AuthService {
	return &authService{
		passwordEncoder: passwordEncoder,
		jwtService:      jwtService,
		userRepository:  userRepository,
	}
}

func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	refreshJwt, err := a.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	id, authorities, err := a.jwtService.ParseAuthToken(ctx, refreshJwt, refreshToken)
	if err != nil {
		return nil, common.NewServiceError(InvalidArgument, err.Error())
	}

	accessJwt, err := a.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	accessToken, err := a.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authService) SignIn(ctx context.Context, signIn SignIn) (*AuthResponse, error) {
	email := common.ToScDf(signIn.Email)

	user, err := a.userRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(InvalidCredentials, "User not found")
	}

	if err := a.checkConfirmed(user); err != nil {
		return nil, err
	}

	if err := a.checkEnabled(user); err != nil {
		return nil, err
	}

	if err := a.checkPassword(user, &signIn); err != nil {
		return nil, err
	}

	authorities, err := a.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return a.createAuthToken(ctx, user.ID, authorities)
}

func (a *authService) checkConfirmed(user *repository.User) error {
	if !user.Confirmed {
		return common.NewServiceError(ErrPermissionDenied, "Account not confirmed")
	}
	return nil
}

func (a *authService) checkEnabled(user *repository.User) error {
	if !user.Enabled {
		return common.NewServiceError(ErrPermissionDenied, "Account not enabled")
	}
	return nil
}

func (a *authService) checkPassword(user *repository.User, signIn *SignIn) error {
	if err := a.passwordEncoder.Compare(signIn.Password, user.Password); err != nil {
		return common.NewServiceError(InvalidCredentials, "Wrong password")
	}
	return nil
}

func (a *authService) createAuthToken(ctx context.Context, id string, authorities []string) (*AuthResponse, error) {
	accessJwt, err := a.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	accessToken, err := a.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	refreshJwt, err := a.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	refreshToken, err := a.jwtService.GenerateAuthToken(refreshJwt, id, authorities)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authService) getAuthorities(ctx context.Context, id string) ([]string, error) {
	userAuthorities, err := a.userRepository.GetUserAuthorities(ctx, id)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	authorities := make([]string, len(userAuthorities))
	for i, saAuthority := range userAuthorities {
		authorities[i] = saAuthority.Authority
	}
	return authorities, nil
}

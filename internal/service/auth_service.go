package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/janobono/go-util/security"
	"net/http"
)

type AuthService interface {
	ChangeEmail(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangeEmail) (*openapi.AuthenticationResponse, error)
	ChangePassword(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangePassword) (*openapi.AuthenticationResponse, error)
	ChangeUserAttributes(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangeUserAttributes) (*openapi.AuthenticationResponse, error)
	Confirm(ctx context.Context, data *openapi.Confirmation) (*openapi.AuthenticationResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*openapi.AuthenticationResponse, error)
	ResendConfirmation(ctx context.Context, data *openapi.ResendConfirmation) error
	ResetPassword(ctx context.Context, data *openapi.ResetPassword) error
	SignIn(ctx context.Context, data *openapi.SignIn) (*openapi.AuthenticationResponse, error)
	SignUp(ctx context.Context, data *openapi.SignUp) (*openapi.AuthenticationResponse, error)
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

func (a *authService) ChangeEmail(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangeEmail) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a *authService) ChangePassword(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangePassword) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a *authService) ChangeUserAttributes(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangeUserAttributes) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a *authService) Confirm(ctx context.Context, confirmation *openapi.Confirmation) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (*openapi.AuthenticationResponse, error) {
	refreshJwt, err := a.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	id, authorities, err := a.jwtService.ParseAuthToken(ctx, refreshJwt, refreshToken)
	if err != nil {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_FIELD), err.Error())
	}

	accessJwt, err := a.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := a.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	return &openapi.AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authService) ResendConfirmation(ctx context.Context, data *openapi.ResendConfirmation) error {
	//TODO implement me
	panic("implement me")
}

func (a *authService) ResetPassword(ctx context.Context, data *openapi.ResetPassword) error {
	//TODO implement me
	panic("implement me")
}

func (a *authService) SignIn(ctx context.Context, data *openapi.SignIn) (*openapi.AuthenticationResponse, error) {
	email := common.ToScDf(data.Email)

	user, err := a.userRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "user not found")
	}

	if err := a.checkConfirmed(user); err != nil {
		return nil, err
	}

	if err := a.checkEnabled(user); err != nil {
		return nil, err
	}

	if err := a.checkPassword(user, data); err != nil {
		return nil, err
	}

	authorities, sErr := a.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, sErr
	}

	return a.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (a *authService) SignUp(ctx context.Context, data *openapi.SignUp) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a *authService) checkConfirmed(user *repository.User) error {
	if !user.Confirmed {
		return common.NewServiceError(http.StatusForbidden, string(openapi.USER_NOT_CONFIRMED), "account not confirmed")
	}
	return nil
}

func (a *authService) checkEnabled(user *repository.User) error {
	if !user.Enabled {
		return common.NewServiceError(http.StatusForbidden, string(openapi.USER_NOT_ENABLED), "account not enabled")
	}
	return nil
}

func (a *authService) checkPassword(user *repository.User, signIn *openapi.SignIn) error {
	if err := a.passwordEncoder.Compare(signIn.Password, user.Password); err != nil {
		return common.NewServiceError(http.StatusForbidden, string(openapi.INVALID_CREDENTIALS), "wrong password")
	}
	return nil
}

func (a *authService) createAuthenticationResponse(ctx context.Context, id pgtype.UUID, authorities []string) (*openapi.AuthenticationResponse, error) {
	accessJwt, err := a.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := a.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	refreshJwt, err := a.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.jwtService.GenerateAuthToken(refreshJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	return &openapi.AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authService) getAuthorities(ctx context.Context, id pgtype.UUID) ([]string, error) {
	userAuthorities, err := a.userRepository.GetUserAuthorities(ctx, id)
	if err != nil {
		return nil, err
	}

	authorities := make([]string, len(userAuthorities))
	for i, saAuthority := range userAuthorities {
		authorities[i] = saAuthority.Authority
	}
	return authorities, nil
}

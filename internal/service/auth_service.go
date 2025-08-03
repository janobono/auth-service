package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/auth-service/internal/service/client"
	"github.com/janobono/go-util/common"
	"github.com/janobono/go-util/security"
	"net/http"
)

type AuthService struct {
	passwordEncoder *security.PasswordEncoder
	captchaClient   client.CaptchaClient
	mailClient      client.MailClient
	jwtService      *JwtService
	userRepository  repository.UserRepository
}

func NewAuthService(
	passwordEncoder *security.PasswordEncoder,
	captchaClient client.CaptchaClient,
	mailClient client.MailClient,
	jwtService *JwtService,
	userRepository repository.UserRepository,
) *AuthService {
	return &AuthService{
		passwordEncoder: passwordEncoder,
		captchaClient:   captchaClient,
		mailClient:      mailClient,
		jwtService:      jwtService,
		userRepository:  userRepository,
	}
}

func (as *AuthService) ChangeEmail(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangeEmail) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (as *AuthService) ChangePassword(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangePassword) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (as *AuthService) ChangeUserAttributes(ctx context.Context, userDetail *openapi.UserDetail, data *openapi.ChangeUserAttributes) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (as *AuthService) Confirm(ctx context.Context, data *openapi.Confirmation) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (as *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*openapi.AuthenticationResponse, error) {
	refreshJwt, err := as.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	id, authorities, err := as.jwtService.ParseAuthToken(ctx, refreshJwt, refreshToken)
	if err != nil {
		return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_FIELD), err.Error())
	}

	accessJwt, err := as.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := as.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	return &openapi.AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) ResendConfirmation(ctx context.Context, data *openapi.ResendConfirmation) error {
	//TODO implement me
	panic("implement me")
}

func (as *AuthService) ResetPassword(ctx context.Context, data *openapi.ResetPassword) error {
	//TODO implement me
	panic("implement me")
}

func (as *AuthService) SignIn(ctx context.Context, data *openapi.SignIn) (*openapi.AuthenticationResponse, error) {
	email := common.ToScDf(data.Email)

	user, err := as.userRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "user not found")
	}

	if err := as.checkConfirmed(user); err != nil {
		return nil, err
	}

	if err := as.checkEnabled(user); err != nil {
		return nil, err
	}

	if err := as.checkPassword(user, data); err != nil {
		return nil, err
	}

	authorities, sErr := as.getAuthorities(ctx, user.ID)
	if err != nil {
		return nil, sErr
	}

	return as.createAuthenticationResponse(ctx, user.ID, authorities)
}

func (as *AuthService) SignUp(ctx context.Context, data *openapi.SignUp) (*openapi.AuthenticationResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (as *AuthService) checkConfirmed(user *repository.User) error {
	if !user.Confirmed {
		return common.NewServiceError(http.StatusForbidden, string(openapi.USER_NOT_CONFIRMED), "account not confirmed")
	}
	return nil
}

func (as *AuthService) checkEnabled(user *repository.User) error {
	if !user.Enabled {
		return common.NewServiceError(http.StatusForbidden, string(openapi.USER_NOT_ENABLED), "account not enabled")
	}
	return nil
}

func (as *AuthService) checkPassword(user *repository.User, signIn *openapi.SignIn) error {
	if err := as.passwordEncoder.Compare(signIn.Password, user.Password); err != nil {
		return common.NewServiceError(http.StatusForbidden, string(openapi.INVALID_CREDENTIALS), "wrong password")
	}
	return nil
}

func (as *AuthService) createAuthenticationResponse(ctx context.Context, id pgtype.UUID, authorities []string) (*openapi.AuthenticationResponse, error) {
	accessJwt, err := as.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := as.jwtService.GenerateAuthToken(accessJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	refreshJwt, err := as.jwtService.GetRefreshJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	refreshToken, err := as.jwtService.GenerateAuthToken(refreshJwt, id, authorities)
	if err != nil {
		return nil, err
	}

	return &openapi.AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) getAuthorities(ctx context.Context, id pgtype.UUID) ([]string, error) {
	userAuthorities, err := as.userRepository.GetUserAuthorities(ctx, id)
	if err != nil {
		return nil, err
	}

	authorities := make([]string, len(userAuthorities))
	for i, saAuthority := range userAuthorities {
		authorities[i] = saAuthority.Authority
	}
	return authorities, nil
}

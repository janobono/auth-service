package impl

import (
	"context"

	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/security"
)

type userDetailDecoder struct {
	jwtService  *service.JwtService
	userService *service.UserService
}

var _ security.UserDetailDecoder[*openapi.UserDetail] = (*userDetailDecoder)(nil)

func NewUserDetailDecoder(jwtService *service.JwtService, userService *service.UserService) security.UserDetailDecoder[*openapi.UserDetail] {
	return &userDetailDecoder{jwtService, userService}
}

func (ud *userDetailDecoder) DecodeGrpcUserDetail(ctx context.Context, token string) (*openapi.UserDetail, error) {
	jwtToken, err := ud.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, err
	}

	id, _, err := ud.jwtService.ParseAuthToken(ctx, jwtToken, token)
	if err != nil {
		return nil, err
	}

	return ud.userService.GetUser(ctx, id)
}

func (ud *userDetailDecoder) GetGrpcUserAuthorities(ctx context.Context, userDetail *openapi.UserDetail) ([]string, error) {
	var authorities = make([]string, len(userDetail.Authorities))
	for i, authority := range userDetail.Authorities {
		authorities[i] = authority.Authority
	}
	return authorities, nil
}

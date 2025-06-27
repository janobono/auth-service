package service

import (
	"context"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/janobono/go-util/security"
)

type userDetailDecoder struct {
	jwtService     *JwtService
	userRepository repository.UserRepository
}

var _ security.UserDetailDecoder[*openapi.UserDetail] = (*userDetailDecoder)(nil)

func NewUserDetailDecoder(jwtService *JwtService, userRepository repository.UserRepository) security.UserDetailDecoder[*openapi.UserDetail] {
	return &userDetailDecoder{jwtService, userRepository}
}

func (ud *userDetailDecoder) DecodeGrpcUserDetail(ctx context.Context, token string) (*openapi.UserDetail, error) {
	jwtToken, err := ud.jwtService.GetAccessJwtToken(ctx)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	id, _, err := ud.jwtService.ParseAuthToken(ctx, jwtToken, token)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	user, err := ud.userRepository.GetUser(ctx, id)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	return mapUserDetail(ctx, ud.userRepository, user)
}

func (ud *userDetailDecoder) GetGrpcUserAuthorities(ctx context.Context, userDetail *openapi.UserDetail) ([]string, error) {
	var authorities = make([]string, len(userDetail.Authorities))
	for i, authority := range userDetail.Authorities {
		authorities[i] = authority.Authority
	}
	return authorities, nil
}

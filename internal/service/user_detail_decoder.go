package service

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/pkg/security"
	"github.com/janobono/auth-service/pkg/util"
)

type userDetailDecoder struct {
	dataSource *db.DataSource
	jwtService JwtService
}

func NewUserDetailDecoder(dataSource *db.DataSource, jwtService JwtService) security.UserDetailDecoder {
	return &userDetailDecoder{dataSource, jwtService}
}

func (ud *userDetailDecoder) DecodeGrpcUserDetail(accessToken string) (*authgrpc.UserDetail, error) {
	jwtToken, err := ud.jwtService.GetAccessJwtToken()
	if err != nil {
		return nil, err
	}

	id, _, err := jwtToken.ParseAccessToken(accessToken)
	if err != nil {
		return nil, err
	}

	uuid, err := util.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	user, err := ud.dataSource.Queries.GetUser(context.Background(), uuid)
	if err != nil {
		return nil, err
	}

	userAttributes, err := ud.dataSource.Queries.GetUserAttributes(context.Background(), user.ID)
	if err != nil {
		return nil, err
	}

	userAuthorities, err := ud.dataSource.Queries.GetUserAuthorities(context.Background(), user.ID)
	if err != nil {
		return nil, err
	}

	var attributes map[string]string
	for _, userAttribute := range userAttributes {
		if !userAttribute.Hidden {
			attributes[userAttribute.Key] = userAttribute.Value.String
		}
	}

	var authorities []string
	for _, userAuthority := range userAuthorities {
		authorities = append(authorities, userAuthority.Authority)
	}

	return &authgrpc.UserDetail{
		Id:          user.ID.String(),
		Email:       user.Email,
		Confirmed:   user.Confirmed,
		Enabled:     user.Enabled,
		Attributes:  attributes,
		Authorities: authorities,
	}, nil
}

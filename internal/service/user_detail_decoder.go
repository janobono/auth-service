package service

import (
	"context"
	"github.com/janobono/auth-service/api"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/pkg/util"
)

type UserDetailDecoder struct {
	dataSource *db.DataSource
	jwtToken   *util.JwtToken
}

func NewUserDetailDecoder(dataSource *db.DataSource, jwtToken *util.JwtToken) *UserDetailDecoder {
	return &UserDetailDecoder{
		dataSource: dataSource,
		jwtToken:   jwtToken,
	}
}

func (ud *UserDetailDecoder) Decode(token string) (*api.UserDetail, error) {
	jwtContent, err := ud.jwtToken.ParseToken(token)
	if err != nil {
		return nil, err
	}

	user, err := ud.dataSource.Queries.GetUser(context.Background(), jwtContent.ID)
	if err != nil {
		return nil, err
	}

	return &api.UserDetail{
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Confirmed:   user.Confirmed,
		Enabled:     user.Enabled,
		Authorities: jwtContent.Authorities,
	}, nil
}

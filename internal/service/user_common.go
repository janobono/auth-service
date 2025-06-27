package service

import (
	"context"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
)

func mapUserDetail(ctx context.Context, userRepository repository.UserRepository, user *repository.User) (*openapi.UserDetail, error) {
	userAttributes, err := userRepository.GetUserAttributes(ctx, user.ID)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	userAuthorities, err := userRepository.GetUserAuthorities(ctx, user.ID)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	attributes := make([]openapi.AttributeValueDetail, 0, len(userAttributes))
	for _, userAttribute := range userAttributes {
		if !userAttribute.Attribute.Hidden {
			attributes = append(attributes, openapi.AttributeValueDetail{
				Key:   userAttribute.Attribute.Key,
				Value: userAttribute.Value,
			})
		}
	}

	authorities := make([]openapi.AuthorityDetail, len(userAuthorities))
	for i, userAuthority := range userAuthorities {
		authorities[i] = openapi.AuthorityDetail{
			Id:        userAuthority.ID.String(),
			Authority: userAuthority.Authority,
		}
	}

	return &openapi.UserDetail{
		Id:          user.ID.String(),
		Email:       user.Email,
		Confirmed:   user.Confirmed,
		Enabled:     user.Enabled,
		Attributes:  attributes,
		Authorities: authorities,
	}, nil
}

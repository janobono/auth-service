package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
)

type UserService interface {
	SearchUsers(ctx context.Context, criteria SearchUsersCriteria, pageable common.Pageable) (*common.Page[*openapi.UserDetail], error)
	GetUser(ctx context.Context, id pgtype.UUID) (*openapi.UserDetail, error)
}

type userService struct {
	userRepository repository.UserRepository
}

var _ UserService = (*userService)(nil)

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{userRepository}
}

func (u *userService) SearchUsers(ctx context.Context, criteria SearchUsersCriteria, pageable common.Pageable) (*common.Page[*openapi.UserDetail], error) {
	page, err := u.userRepository.SearchUsers(ctx, repository.SearchUsersCriteria{
		SearchField:   criteria.SearchField,
		Email:         criteria.Email,
		AttributeKeys: criteria.AttributeKeys,
	}, pageable)
	if err != nil {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}

	content := make([]*openapi.UserDetail, len(page.Content))
	for i, user := range page.Content {
		userDetail, subErr := u.mapUserDetail(ctx, u.userRepository, user)
		if subErr != nil {
			return nil, subErr
		}
		content[i] = userDetail
	}

	return &common.Page[*openapi.UserDetail]{
		Pageable:      &pageable,
		TotalElements: page.TotalElements,
		TotalPages:    page.TotalPages,
		First:         page.First,
		Last:          page.Last,
		Content:       content,
		Empty:         page.Empty,
	}, nil
}

func (u *userService) GetUser(ctx context.Context, id pgtype.UUID) (*openapi.UserDetail, error) {
	user, err := u.userRepository.GetUser(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(string(openapi.UNKNOWN), err.Error())
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(string(openapi.NOT_FOUND), "User not found")
	}
	return u.mapUserDetail(ctx, u.userRepository, user)
}

func (u *userService) mapUserDetail(ctx context.Context, userRepository repository.UserRepository, user *repository.User) (*openapi.UserDetail, error) {
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

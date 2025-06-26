package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
)

type UserService interface {
	SearchUsers(ctx context.Context, criteria SearchUsersCriteria, pageable common.Pageable) (*common.Page[*UserDetail], error)
	GetUser(ctx context.Context, id string) (*UserDetail, error)
}

type userService struct {
	userRepository repository.UserRepository
}

var _ UserService = (*userService)(nil)

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{userRepository}
}

func (u *userService) SearchUsers(ctx context.Context, criteria SearchUsersCriteria, pageable common.Pageable) (*common.Page[*UserDetail], error) {
	page, err := u.userRepository.SearchUsers(ctx, repository.SearchUsersCriteria{
		SearchField:   criteria.SearchField,
		Email:         criteria.Email,
		AttributeKeys: criteria.AttributeKeys,
	}, pageable)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	content := make([]*UserDetail, len(page.Content))
	for i, user := range page.Content {
		userDetail, err := u.mapUserDetail(ctx, user)
		if err != nil {
			return nil, common.NewServiceError(ErrInternalError, err.Error())
		}
		content[i] = userDetail
	}

	return &common.Page[*UserDetail]{
		Pageable:      &pageable,
		TotalElements: page.TotalElements,
		TotalPages:    page.TotalPages,
		First:         page.First,
		Last:          page.Last,
		Content:       content,
		Empty:         page.Empty,
	}, nil
}

func (u *userService) GetUser(ctx context.Context, id string) (*UserDetail, error) {
	user, err := u.userRepository.GetUser(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(ErrNotFound, "User not found")
	}
	return u.mapUserDetail(ctx, user)
}

func (u *userService) mapUserDetail(ctx context.Context, user *repository.User) (*UserDetail, error) {
	userAttributes, err := u.userRepository.GetUserAttributes(ctx, user.ID)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	userAuthorities, err := u.userRepository.GetUserAuthorities(ctx, user.ID)
	if err != nil {
		return nil, common.NewServiceError(ErrInternalError, err.Error())
	}

	attributes := make([]*UserAttribute, 0, len(userAttributes))
	for _, userAttribute := range userAttributes {
		if !userAttribute.Attribute.Hidden {
			attributes = append(attributes, &UserAttribute{
				Attribute: &Attribute{
					Id:       userAttribute.Attribute.ID,
					Key:      userAttribute.Attribute.Key,
					Name:     userAttribute.Attribute.Name,
					Required: userAttribute.Attribute.Required,
					Hidden:   userAttribute.Attribute.Hidden,
				},
				Value: userAttribute.Value,
			})
		}
	}

	authorities := make([]*Authority, len(userAuthorities))
	for i, userAuthority := range userAuthorities {
		authorities[i] = &Authority{
			Id:        userAuthority.ID,
			Authority: userAuthority.Authority,
		}
	}

	return &UserDetail{
		Id:          user.ID,
		Email:       user.Email,
		Confirmed:   user.Confirmed,
		Enabled:     user.Enabled,
		Attributes:  attributes,
		Authorities: authorities,
	}, nil
}

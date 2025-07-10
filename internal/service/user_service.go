package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"net/http"
)

type UserService interface {
	AddUser(ctx context.Context, userData *openapi.UserData) (*openapi.UserDetail, error)
	DeleteUser(ctx context.Context, id pgtype.UUID) error
	GetUser(ctx context.Context, id pgtype.UUID) (*openapi.UserDetail, error)
	GetUsers(ctx context.Context, criteria *SearchUserCriteria, pageable *common.Pageable) (*common.Page[*openapi.UserDetail], error)
	SetAuthorities(ctx context.Context, id pgtype.UUID, userAuthoritiesData *openapi.UserAuthoritiesData) (*openapi.UserDetail, error)
	SetConfirmed(ctx context.Context, id pgtype.UUID, booleanValue *openapi.BooleanValue) (*openapi.UserDetail, error)
	SetEnabled(ctx context.Context, id pgtype.UUID, booleanValue *openapi.BooleanValue) (*openapi.UserDetail, error)
	SetUser(ctx context.Context, id pgtype.UUID, userData *openapi.UserData) (*openapi.UserDetail, error)
}

type userService struct {
	userRepository repository.UserRepository
}

var _ UserService = (*userService)(nil)

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{userRepository}
}

func (u *userService) AddUser(ctx context.Context, userData *openapi.UserData) (*openapi.UserDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userService) DeleteUser(ctx context.Context, id pgtype.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (u *userService) GetUser(ctx context.Context, id pgtype.UUID) (*openapi.UserDetail, error) {
	user, err := u.userRepository.GetUser(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "User not found")
	}
	return u.mapUserDetail(ctx, user)
}

func (u *userService) GetUsers(ctx context.Context, criteria *SearchUserCriteria, pageable *common.Pageable) (*common.Page[*openapi.UserDetail], error) {
	page, err := u.userRepository.SearchUsers(ctx, &repository.SearchUsersCriteria{
		SearchField:   criteria.SearchField,
		Email:         criteria.Email,
		AttributeKeys: criteria.AttributeKeys,
	}, pageable)
	if err != nil {
		return nil, err
	}

	content := make([]*openapi.UserDetail, len(page.Content))
	for i, user := range page.Content {
		userDetail, subErr := u.mapUserDetail(ctx, user)
		if subErr != nil {
			return nil, subErr
		}
		content[i] = userDetail
	}

	return &common.Page[*openapi.UserDetail]{
		Pageable:      pageable,
		TotalElements: page.TotalElements,
		TotalPages:    page.TotalPages,
		First:         page.First,
		Last:          page.Last,
		Content:       content,
		Empty:         page.Empty,
	}, nil
}

func (u *userService) SetAuthorities(ctx context.Context, id pgtype.UUID, userAuthoritiesData *openapi.UserAuthoritiesData) (*openapi.UserDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userService) SetConfirmed(ctx context.Context, id pgtype.UUID, booleanValue *openapi.BooleanValue) (*openapi.UserDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userService) SetEnabled(ctx context.Context, id pgtype.UUID, booleanValue *openapi.BooleanValue) (*openapi.UserDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userService) SetUser(ctx context.Context, id pgtype.UUID, userData *openapi.UserData) (*openapi.UserDetail, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userService) mapUserDetail(ctx context.Context, user *repository.User) (*openapi.UserDetail, error) {
	userAttributes, err := u.userRepository.GetUserAttributes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	userAuthorities, err := u.userRepository.GetUserAuthorities(ctx, user.ID)
	if err != nil {
		return nil, err
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

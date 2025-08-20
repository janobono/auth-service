package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/go-util/common"
	"github.com/janobono/go-util/security"
)

type UserService struct {
	passwordEncoder     *security.PasswordEncoder
	randomString        *security.RandomString
	attributeRepository repository.AttributeRepository
	authorityRepository repository.AuthorityRepository
	userRepository      repository.UserRepository
}

func NewUserService(
	passwordEncoder *security.PasswordEncoder,
	randomString *security.RandomString,
	attributeRepository repository.AttributeRepository,
	authorityRepository repository.AuthorityRepository,
	userRepository repository.UserRepository,
) *UserService {
	return &UserService{
		passwordEncoder,
		randomString,
		attributeRepository,
		authorityRepository,
		userRepository,
	}
}

func (u *UserService) AddUser(ctx context.Context, data *openapi.UserData) (*openapi.UserDetail, error) {
	email := common.ToScDf(data.Email)

	count, err := u.userRepository.CountByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.EMAIL_ALREADY_EXISTS), "'email' already exists")
	}

	password, err := u.randomString.Generate()
	if err != nil {
		return nil, err
	}

	password, err = u.passwordEncoder.Encode(password)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepository.AddUser(ctx, &repository.UserData{
		Email:     email,
		Password:  password,
		Confirmed: data.Confirmed,
		Enabled:   data.Enabled,
	})
	if err != nil {
		return nil, err
	}

	return u.mapUserDetail(ctx, user)
}

func (u *UserService) DeleteUser(ctx context.Context, userDetail *openapi.UserDetail, id pgtype.UUID) error {
	err := u.checkUser(ctx, userDetail, id)
	if err != nil {
		return err
	}

	return u.userRepository.DeleteUserById(ctx, id)
}

func (u *UserService) GetUser(ctx context.Context, id pgtype.UUID) (*openapi.UserDetail, error) {
	user, err := u.userRepository.GetUserById(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "User not found")
	}
	return u.mapUserDetail(ctx, user)
}

func (u *UserService) GetUsers(ctx context.Context, criteria *SearchUserCriteria, pageable *common.Pageable) (*common.Page[*openapi.UserDetail], error) {
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

func (u *UserService) SetAttributes(ctx context.Context, userDetail *openapi.UserDetail, id pgtype.UUID, data *openapi.UserAttributesData) (*openapi.UserDetail, error) {
	err := u.checkUser(ctx, userDetail, id)
	if err != nil {
		return nil, err
	}

	userAttributeMap := make(map[string]string, len(data.Attributes))
	for _, userAttribute := range data.Attributes {
		userAttributeMap[userAttribute.Key] = userAttribute.Value
	}

	allAttributes, err := u.attributeRepository.GetAllAttributes(ctx)
	if err != nil {
		return nil, err
	}

	userAttributes := make([]*repository.UserAttribute, 0, len(userAttributeMap))
	for _, attribute := range allAttributes {
		value, ok := userAttributeMap[attribute.Key]

		if attribute.Required && !ok {
			return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.REQUIRED_ATTRIBUTE), fmt.Sprintf("attribute %s is required", attribute.Key))
		}

		if ok {
			if common.IsBlank(value) {
				return nil, common.NewServiceError(http.StatusBadRequest, string(openapi.INVALID_FIELD), "'value' must not be blank")
			}

			userAttributes = append(userAttributes, &repository.UserAttribute{
				Attribute: attribute,
				Value:     value,
			})
		}
	}

	_, err = u.userRepository.SetUserAttributes(ctx, &repository.UserAttributesData{
		UserID:     id,
		Attributes: userAttributes,
	})
	if err != nil {
		return nil, err
	}

	return u.GetUser(ctx, id)
}

func (u *UserService) SetAuthorities(ctx context.Context, userDetail *openapi.UserDetail, id pgtype.UUID, data *openapi.UserAuthoritiesData) (*openapi.UserDetail, error) {
	err := u.checkUser(ctx, userDetail, id)
	if err != nil {
		return nil, err
	}

	userAuthorities := make([]*repository.Authority, 0, len(data.Authorities))
	for _, userAuthority := range data.Authorities {
		authority, err := u.authorityRepository.GetAuthorityByAuthority(ctx, userAuthority)
		if err != nil {
			return nil, err
		}

		userAuthorities = append(userAuthorities, authority)
	}

	_, err = u.userRepository.SetUserAuthorities(ctx, &repository.UserAuthoritiesData{
		UserID:      id,
		Authorities: userAuthorities,
	})
	if err != nil {
		return nil, err
	}

	return u.GetUser(ctx, id)
}

func (u *UserService) SetConfirmed(ctx context.Context, userDetail *openapi.UserDetail, id pgtype.UUID, data *openapi.BooleanValue) (*openapi.UserDetail, error) {
	err := u.checkUser(ctx, userDetail, id)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepository.SetUserConfirmed(ctx, id, data.Value)
	if err != nil {
		return nil, err
	}

	return u.mapUserDetail(ctx, user)
}

func (u *UserService) SetEmail(ctx context.Context, userDetail *openapi.UserDetail, id pgtype.UUID, data *openapi.UserEmailData) (*openapi.UserDetail, error) {
	err := u.checkUser(ctx, userDetail, id)
	if err != nil {
		return nil, err
	}

	email := common.ToScDf(data.Email)

	count, err := u.userRepository.CountByEmailAndNotId(ctx, email, id)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, common.NewServiceError(http.StatusConflict, string(openapi.INVALID_FIELD), "'email' already exists")
	}

	user, err := u.userRepository.SetUserEmail(ctx, id, email)
	if err != nil {
		return nil, err
	}

	return u.mapUserDetail(ctx, user)
}

func (u *UserService) SetEnabled(ctx context.Context, userDetail *openapi.UserDetail, id pgtype.UUID, data *openapi.BooleanValue) (*openapi.UserDetail, error) {
	err := u.checkUser(ctx, userDetail, id)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepository.SetUserEnabled(ctx, id, data.Value)
	if err != nil {
		return nil, err
	}

	return u.mapUserDetail(ctx, user)
}

func (u *UserService) checkUser(ctx context.Context, userDetail *openapi.UserDetail, id pgtype.UUID) error {
	count, err := u.userRepository.CountById(ctx, id)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NewServiceError(http.StatusNotFound, string(openapi.NOT_FOUND), "user does not exist")
	}

	if userDetail.Id == id.String() {
		return common.NewServiceError(http.StatusBadRequest, string(openapi.CANNOT_MANAGE_OWN_ACCOUNT), "cannot manage own account")
	}
	return nil
}

func (u *UserService) mapUserDetail(ctx context.Context, user *repository.User) (*openapi.UserDetail, error) {
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

package service

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/gen/db/repository"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/db/dal"
	"github.com/janobono/auth-service/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type userServer struct {
	authgrpc.UnimplementedUserServer
	dataSource      *db.DataSource
	jwtService      JwtService
	passwordEncoder component.PasswordEncoder
}

func NewUserServer(dataSource *db.DataSource, jwtService JwtService, passwordEncoder component.PasswordEncoder) authgrpc.UserServer {
	return &userServer{
		dataSource:      dataSource,
		jwtService:      jwtService,
		passwordEncoder: passwordEncoder,
	}
}

func (us *userServer) SearchUsers(ctx context.Context, searchCriteria *authgrpc.SearchCriteria) (*authgrpc.UserPage, error) {
	searchUsersParams := toSearchParams(searchCriteria)

	count, err := us.dataSource.DalQueries.CountUsersByCriteria(ctx, *searchUsersParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Count users failed")
	}

	var content []*authgrpc.UserDetail

	if count > 0 {
		users, err := us.dataSource.DalQueries.SearchUsersByCriteria(ctx, *searchUsersParams)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Search users failed")
		}
		for _, user := range users {
			userDetail, err := us.getUserDetail(ctx, &user)
			if err != nil {
				return nil, err
			}
			content = append(content, userDetail)
		}
	}

	return &authgrpc.UserPage{
		Page: &authgrpc.PageDetail{
			Page:          searchUsersParams.Page,
			Size:          searchUsersParams.Size,
			Sort:          sort(searchCriteria),
			TotalPages:    util.TotalPages(searchUsersParams.Size, count),
			TotalElements: count,
		},
		Content: content,
	}, nil
}

func (us *userServer) GetUser(ctx context.Context, id *wrapperspb.StringValue) (*authgrpc.UserDetail, error) {
	uuid, err := util.ParseUUID(id.Value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid id")
	}

	user, err := us.dataSource.Queries.GetUser(ctx, uuid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Get user failed")
	}

	return us.getUserDetail(ctx, &user)
}

func (us *userServer) getUserDetail(ctx context.Context, user *repository.User) (*authgrpc.UserDetail, error) {
	userAttributes, err := us.dataSource.Queries.GetUserAttributes(ctx, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Get user attributes failed")
	}

	userAuthorities, err := us.dataSource.Queries.GetUserAuthorities(ctx, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Get user authorities failed")
	}

	var attributes map[string]string
	for _, userAttribute := range userAttributes {
		attributes[userAttribute.Key] = userAttribute.Value.String
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

func toSearchParams(searchCriteria *authgrpc.SearchCriteria) *dal.SearchUsersParams {
	result := &dal.SearchUsersParams{
		Page: 0,
		Size: 20,
		Sort: sort(searchCriteria),
	}

	if searchCriteria == nil {
		return result
	}

	result.SearchField = searchCriteria.SearchField
	result.Email = searchCriteria.Email
	result.AttributeKeys = searchCriteria.AttributeKeys

	if searchCriteria.Page == nil {
		return result
	}

	result.Page = util.AbsInt32(searchCriteria.Page.Page)
	result.Size = util.AbsInt32(searchCriteria.Page.Size)

	return result
}

func sort(searchCriteria *authgrpc.SearchCriteria) string {
	if searchCriteria == nil || searchCriteria.Page == nil || util.IsBlank(searchCriteria.Page.Sort) {
		return "id asc"
	}

	return searchCriteria.Page.Sort
}

package service

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/gen/db/repository"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/db"
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
	// TODO : Implement
	return nil, status.Errorf(codes.Unimplemented, "Unimplemented")
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

	return us.getUserDetail(&user)
}

func (us *userServer) getUserDetail(user *repository.User) (*authgrpc.UserDetail, error) {
	userAttributes, err := us.dataSource.Queries.GetUserAttributes(context.Background(), user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Get user attributes failed")
	}

	userAuthorities, err := us.dataSource.Queries.GetUserAuthorities(context.Background(), user.ID)
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

package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/gen/db/repository"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type authServer struct {
	authgrpc.UnimplementedAuthServer
	dataSource      *db.DataSource
	jwtService      *JwtService
	passwordEncoder *component.PasswordEncoder
}

func NewAuthServer(dataSource *db.DataSource, jwtService *JwtService, passwordEncoder *component.PasswordEncoder) authgrpc.AuthServer {
	return &authServer{
		dataSource:      dataSource,
		jwtService:      jwtService,
		passwordEncoder: passwordEncoder,
	}
}

func (as *authServer) GetUser(ctx context.Context, empty *emptypb.Empty) (*authgrpc.UserDetail, error) {
	userDetail := GetGrpcUserDetail(ctx)
	if userDetail == nil {
		return nil, status.Errorf(codes.Internal, "Empty context")
	}
	return userDetail, nil
}

func (as *authServer) Refresh(ctx context.Context, refreshToken *wrapperspb.StringValue) (*authgrpc.AuthResponse, error) {
	jwtToken, err := as.jwtService.GetRefreshJwtToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get refresh token")
	}

	id, authorities, err := jwtToken.ParseAccessToken(refreshToken.Value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid refresh token")
	}

	jwtToken, err = as.jwtService.GetAccessJwtToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get access token")
	}

	accessToken, err := jwtToken.GenerateAccessToken(id, *authorities)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate access token")
	}

	return &authgrpc.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Value,
	}, nil
}

func (as *authServer) SignIn(ctx context.Context, signInData *authgrpc.SignInData) (*authgrpc.AuthResponse, error) {
	email := util.ToScDf(signInData.Email)

	user, err := as.dataSource.Queries.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, "Get user failed")
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid credentials")
	}

	err = as.checkConfirmed(&user)
	if err != nil {
		return nil, err
	}

	err = as.checkEnabled(&user)
	if err != nil {
		return nil, err
	}

	err = as.checkPassword(&user, signInData)
	if err != nil {
		return nil, err
	}

	authorities, err := as.getAuthorities(ctx, user.ID)

	return as.createAuthToken(user.ID.String(), authorities)
}

func (as *authServer) checkConfirmed(user *repository.User) error {
	if !user.Confirmed {
		return status.Errorf(codes.PermissionDenied, "Account not confirmed")
	}
	return nil
}

func (as *authServer) checkEnabled(user *repository.User) error {
	if !user.Enabled {
		return status.Errorf(codes.PermissionDenied, "Account not enabled")
	}
	return nil
}

func (as *authServer) checkPassword(user *repository.User, signInData *authgrpc.SignInData) error {
	if err := as.passwordEncoder.Compare(signInData.Password, user.Password); err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid credentials")
	}
	return nil
}

func (as *authServer) createAuthToken(id string, authorities *[]string) (*authgrpc.AuthResponse, error) {
	jwtToken, err := as.jwtService.GetAccessJwtToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get access token")
	}

	accessToken, err := jwtToken.GenerateAccessToken(id, *authorities)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate access token")
	}

	jwtToken, err = as.jwtService.GetRefreshJwtToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get refresh token")
	}

	refreshToken, err := jwtToken.GenerateAccessToken(id, *authorities)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate refresh token")
	}

	return &authgrpc.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *authServer) getAuthorities(ctx context.Context, id pgtype.UUID) (*[]string, error) {
	saAuthorities, err := as.dataSource.Queries.GetUserAuthorities(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user authorities")
	}

	authorities := make([]string, len(saAuthorities))
	for i, saAuthority := range saAuthorities {
		authorities[i] = saAuthority.Authority
	}
	return &authorities, nil
}

package impl

import (
	"context"
	"github.com/janobono/auth-service/generated/openapi"
	"github.com/janobono/auth-service/generated/proto"
	"github.com/janobono/auth-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log/slog"

	"github.com/janobono/go-util/common"
	"github.com/janobono/go-util/security"
)

type authServer struct {
	proto.UnimplementedAuthServer
	authService service.AuthService
}

var _ proto.AuthServer = (*authServer)(nil)

func NewAuthServer(authService service.AuthService) proto.AuthServer {
	return &authServer{authService: authService}
}

func (as *authServer) GetUser(ctx context.Context, empty *emptypb.Empty) (*proto.UserDetail, error) {
	userDetail, ok := security.GetGrpcUserDetail[*proto.UserDetail](ctx)
	if userDetail == nil || !ok {
		slog.Error("Empty ok invalid context")
		return nil, status.Errorf(codes.Unauthenticated, "%s", "Empty or invalid context")
	}
	return userDetail, nil
}

func (as *authServer) Refresh(ctx context.Context, refreshToken *wrapperspb.StringValue) (*proto.AuthResponse, error) {
	authenticationResponse, err := as.authService.RefreshToken(ctx, refreshToken.Value)
	if err != nil {
		slog.Error("RefreshToken failed", "error", err)
		switch {
		case common.IsCode(err, string(openapi.INVALID_FIELD)):
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	return &proto.AuthResponse{
		AccessToken:  authenticationResponse.AccessToken,
		RefreshToken: authenticationResponse.RefreshToken,
	}, nil
}

func (as *authServer) SignIn(ctx context.Context, signInData *proto.SignInData) (*proto.AuthResponse, error) {
	authenticationResponse, err := as.authService.SignIn(ctx, &openapi.SignIn{Email: signInData.Email, Password: signInData.Password})
	if err != nil {
		slog.Error("SignIn failed", "error", err)
		switch {
		case common.IsCode(err, string(openapi.NOT_FOUND)):
			return nil, status.Errorf(codes.NotFound, "%s", err.Error())
		case common.IsCode(err, string(openapi.INVALID_CREDENTIALS)):
		case common.IsCode(err, string(openapi.USER_NOT_ENABLED)):
		case common.IsCode(err, string(openapi.USER_NOT_CONFIRMED)):
			return nil, status.Errorf(codes.PermissionDenied, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	return &proto.AuthResponse{
		AccessToken:  authenticationResponse.AccessToken,
		RefreshToken: authenticationResponse.RefreshToken,
	}, nil
}

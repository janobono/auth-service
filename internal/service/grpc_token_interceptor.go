package service

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/pkg/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

const userDetailKey = "userDetail"
const bearerPrefix = "Bearer "

type GrpcTokenInterceptor struct {
	userDetailDecoder *UserDetailDecoder
}

func NewGrpcTokenInterceptor(userDetailDecoder *UserDetailDecoder) *GrpcTokenInterceptor {
	return &GrpcTokenInterceptor{userDetailDecoder}
}

func (g *GrpcTokenInterceptor) InterceptToken(methods *[]security.GrpcSecuredMethod) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		securedMethod := security.FindGrpcSecuredMethod(methods, info.FullMethod)

		if securedMethod != nil {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
			}

			authHeader := md["authorization"]
			if len(authHeader) == 0 || !strings.HasPrefix(authHeader[0], bearerPrefix) {
				return nil, status.Errorf(codes.Unauthenticated, "missing or invalid Bearer token")
			}

			token := authHeader[0][len(bearerPrefix):]
			userDetail, err := g.userDetailDecoder.DecodeGrpcUserDetail(token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}

			if len(securedMethod.Authorities) > 0 && !security.HasAnyAuthority(&securedMethod.Authorities, &userDetail.Authorities) {
				return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
			}

			ctx = context.WithValue(ctx, userDetailKey, userDetail)
		}

		return handler(ctx, req)
	}
}

func GetGrpcUserDetail(ctx context.Context) *authgrpc.UserDetail {
	value := ctx.Value(userDetailKey)
	if value == nil {
		return nil
	}
	return value.(*authgrpc.UserDetail)
}

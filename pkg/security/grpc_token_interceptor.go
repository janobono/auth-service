package security

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type grpcTokenInterceptor struct {
	userDetailDecoder UserDetailDecoder
}

func NewGrpcTokenInterceptor(userDetailDecoder UserDetailDecoder) GrpcTokenInterceptor {
	return &grpcTokenInterceptor{userDetailDecoder}
}

func (g *grpcTokenInterceptor) InterceptToken(methods *[]GrpcSecuredMethod) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		securedMethod := findGrpcSecuredMethod(methods, info.FullMethod)

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

			if len(securedMethod.Authorities) > 0 && !HasAnyAuthority(&securedMethod.Authorities, &userDetail.Authorities) {
				return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
			}

			ctx = context.WithValue(ctx, userDetailKey, userDetail)
		}

		return handler(ctx, req)
	}
}

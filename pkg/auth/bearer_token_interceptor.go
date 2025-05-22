package auth

import (
	"context"
	"github.com/janobono/auth-service/api"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const userDetailKey = "userDetail"
const bearerPrefix = "Bearer "

type SecuredMethod struct {
	Method      string
	Authorities []string
}

type UserDetailDecoder interface {
	Decode(token string) (*api.UserDetail, error)
}

type BearerTokenInterceptor struct {
	UserDetailDecoder UserDetailDecoder
}

func NewBearerTokenInterceptor(userDetailDecoder UserDetailDecoder) *BearerTokenInterceptor {
	return &BearerTokenInterceptor{
		UserDetailDecoder: userDetailDecoder,
	}
}

func (bti *BearerTokenInterceptor) CheckToken(secureMethods *[]SecuredMethod) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		securedMethod := findSecuredMethod(secureMethods, info.FullMethod)

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
			userDetail, err := bti.UserDetailDecoder.Decode(token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}

			if len(securedMethod.Authorities) > 0 && !hasAnyAuthority(&securedMethod.Authorities, &userDetail.Authorities) {
				return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
			}

			ctx = context.WithValue(ctx, userDetailKey, userDetail)
		}

		return handler(ctx, req)
	}
}

func GetUserDetail(ctx context.Context) *api.UserDetail {
	return ctx.Value(userDetailKey).(*api.UserDetail)
}

func findSecuredMethod(methods *[]SecuredMethod, methodName string) *SecuredMethod {
	for _, method := range *methods {
		if method.Method == methodName {
			return &method
		}
	}
	return nil
}

func hasAnyAuthority(methodAuthorities, userAuthorities *[]string) bool {
	set := make(map[string]bool)

	for _, item := range *methodAuthorities {
		set[item] = true
	}

	for _, item := range *userAuthorities {
		if _, found := set[item]; found {
			return true
		}
	}

	return false
}

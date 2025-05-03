package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const UserDetailKey = "userDetail"
const bearerPrefix = "Bearer "

type UserDetail struct {
	Id          int64
	Email       string
	FirstName   string
	LastName    string
	Confirmed   bool
	Enabled     bool
	Authorities []string
}

type SecuredMethod struct {
	Method      string
	Authorities []string
}

type TokenDecoder interface {
	DecodeToken(token string) (UserDetail, error)
}

type AuthorizationInterceptor struct {
	TokenDecoder TokenDecoder
}

func (interceptor *AuthorizationInterceptor) CheckToken(secureMethods *[]SecuredMethod) grpc.UnaryServerInterceptor {
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
			userDetail, err := interceptor.TokenDecoder.DecodeToken(token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}

			if len(securedMethod.Authorities) > 0 && !hasAnyAuthority(securedMethod.Authorities, userDetail.Authorities) {
				return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
			}

			ctx = context.WithValue(ctx, UserDetailKey, userDetail)
		}

		return handler(ctx, req)
	}
}

func findSecuredMethod(methods *[]SecuredMethod, methodName string) *SecuredMethod {
	for _, method := range *methods {
		if method.Method == methodName {
			return &method
		}
	}
	return nil
}

func hasAnyAuthority(methodAuthorities, userAuthorities []string) bool {
	set := make(map[string]bool)

	for _, item := range methodAuthorities {
		set[item] = true
	}

	for _, item := range userAuthorities {
		if _, found := set[item]; found {
			return true
		}
	}

	return false
}

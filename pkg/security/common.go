package security

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
)

const userDetailKey = "userDetail"
const bearerPrefix = "Bearer "

func GetGrpcUserDetail(ctx context.Context) *authgrpc.UserDetail {
	value := ctx.Value(userDetailKey)
	if value == nil {
		return nil
	}
	return value.(*authgrpc.UserDetail)
}

func HasAnyAuthority(methodAuthorities, userAuthorities *[]string) bool {
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

func findGrpcSecuredMethod(methods *[]GrpcSecuredMethod, methodName string) *GrpcSecuredMethod {
	for _, method := range *methods {
		if method.Method == methodName {
			return &method
		}
	}
	return nil
}

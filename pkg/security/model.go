package security

import (
	"github.com/janobono/auth-service/gen/authgrpc"
	"google.golang.org/grpc"
)

type GrpcSecuredMethod struct {
	Method      string
	Authorities []string
}

type UserDetailDecoder interface {
	DecodeGrpcUserDetail(accessToken string) (*authgrpc.UserDetail, error)
}

type GrpcTokenInterceptor interface {
	InterceptToken(secureMethods *[]GrpcSecuredMethod) grpc.UnaryServerInterceptor
}

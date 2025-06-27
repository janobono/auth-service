package server

import (
	"github.com/janobono/auth-service/generated/proto"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/server/impl"
	"github.com/janobono/go-util/security"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type GrpcServer struct {
	config   *config.ServerConfig
	services *Services
}

func NewGrpcServer(config *config.ServerConfig, services *Services) *GrpcServer {
	return &GrpcServer{config, services}
}

func (s *GrpcServer) Start() *grpc.Server {
	slog.Info("Starting gRPC server...")

	lis, err := net.Listen("tcp", s.config.GRPCAddress)
	if err != nil {
		slog.Error("Failed to listen", "error", err)
		panic(err)
	}

	grpcTokenInterceptor := security.NewGrpcTokenInterceptor(impl.NewUserDetailDecoder(s.services.JwtService, s.services.UserService)).InterceptToken(
		[]security.GrpcSecuredMethod{
			{
				Method:      proto.User_SearchUsers_FullMethodName,
				Authorities: append(s.config.SecurityConfig.ReadAuthorities, s.config.SecurityConfig.WriteAuthorities...),
			},
			{
				Method:      proto.User_GetUser_FullMethodName,
				Authorities: append(s.config.SecurityConfig.ReadAuthorities, s.config.SecurityConfig.WriteAuthorities...),
			},
		})

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcTokenInterceptor))

	proto.RegisterAuthServer(grpcServer, impl.NewAuthServer(s.services.AuthService))
	proto.RegisterUserServer(grpcServer, impl.NewUserServer(s.services.UserService))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Failed to serve", "error", err)
			panic(err)
		}
	}()

	slog.Info("gRPC server started", "port", s.config.GRPCAddress)
	return grpcServer
}

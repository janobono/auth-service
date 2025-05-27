package server

import (
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/auth-service/pkg/security"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type GrpcServer struct {
	config            *config.ServerConfig
	dataSource        *db.DataSource
	jwtService        *service.JwtService
	userDetailDecoder *service.UserDetailDecoder
	passwordEncoder   *component.PasswordEncoder
}

func NewGrpcServer(
	config *config.ServerConfig,
	dataSource *db.DataSource,
	jwtService *service.JwtService,
	userDetailDecoder *service.UserDetailDecoder,
	passwordEncoder *component.PasswordEncoder,
) *GrpcServer {
	return &GrpcServer{config, dataSource, jwtService, userDetailDecoder, passwordEncoder}
}

func (s *GrpcServer) Start() *grpc.Server {
	slog.Info("Starting gRPC server...")

	lis, err := net.Listen("tcp", s.config.GRPCAddress)
	if err != nil {
		slog.Error("Failed to listen", "error", err)
		panic(err)
	}

	grpcTokenInterceptor := service.NewGrpcTokenInterceptor(s.userDetailDecoder).InterceptToken(&[]security.GrpcSecuredMethod{
		{
			Method:      authgrpc.Captcha_IsValid_FullMethodName,
			Authorities: []string{},
		},
		{
			Method:      authgrpc.User_SearchUsers_FullMethodName,
			Authorities: []string{s.config.SecurityConfig.AuthorityAdmin, s.config.SecurityConfig.AuthorityManager},
		},
		{
			Method:      authgrpc.User_GetUser_FullMethodName,
			Authorities: []string{s.config.SecurityConfig.AuthorityAdmin, s.config.SecurityConfig.AuthorityManager},
		},
	})

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcTokenInterceptor))

	authgrpc.RegisterAuthServer(grpcServer, service.NewAuthServer(s.dataSource, s.jwtService, s.passwordEncoder))
	authgrpc.RegisterCaptchaServer(grpcServer, service.NewCaptchaServer(s.passwordEncoder))
	authgrpc.RegisterUserServer(grpcServer, service.NewUserServer(s.dataSource, s.jwtService, s.passwordEncoder))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Failed to serve", "error", err)
			panic(err)
		}
	}()

	slog.Info("gRPC server started", "port", s.config.GRPCAddress)
	return grpcServer
}

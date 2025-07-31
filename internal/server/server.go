package server

import (
	"context"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/repository"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/go-util/security"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Services struct {
	AttributeService service.AttributeService
	AuthService      service.AuthService
	AuthorityService service.AuthorityService
	JwkService       service.JwkService
	JwtService       *service.JwtService
	UserService      service.UserService
}

type Server struct {
	config *config.ServerConfig
}

func NewServer(config *config.ServerConfig) *Server {
	initSlog(config)
	return &Server{config}
}

func (s *Server) Start() {
	slog.Info("Starting server...")

	dataSource := db.NewDataSource(s.config.DbConfig)
	defer dataSource.Close()

	initDefaultCredentials(s.config, dataSource)

	passwordEncoder := security.NewPasswordEncoder(bcrypt.DefaultCost)
	randomString := security.NewRandomString(s.config.AppConfig.PasswordCharacters, s.config.AppConfig.PasswordLength)

	attributeRepository := repository.NewAttributeRepository(dataSource)
	authorityRepository := repository.NewAuthorityRepository(dataSource)
	jwkRepository := repository.NewJwkRepository(dataSource)
	userRepository := repository.NewUserRepository(dataSource)

	jwtService := service.NewJwtService(s.config.SecurityConfig, jwkRepository)

	services := &Services{
		AttributeService: service.NewAttributeService(attributeRepository),
		AuthService:      service.NewAuthService(passwordEncoder, jwtService, userRepository),
		AuthorityService: service.NewAuthorityService(authorityRepository),
		JwkService:       service.NewJwkService(jwkRepository),
		JwtService:       service.NewJwtService(s.config.SecurityConfig, jwkRepository),
		UserService: service.NewUserService(
			passwordEncoder,
			randomString,
			attributeRepository,
			authorityRepository,
			userRepository,
		),
	}

	grpcServer := NewGrpcServer(s.config, services).Start()
	httpServer := NewHttpServer(s.config, services).Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	slog.Info("Server started. Press Ctrl+C to exit.")

	<-stop
	slog.Info("Shutting down server...")

	grpcServer.GracefulStop()
	slog.Info("gRPC server stopped gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("Http server forced to stop", "error", err)
	} else {
		slog.Info("Http server stopped gracefully")
	}

	slog.Info("Server shut down")
}

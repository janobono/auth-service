package server

import (
	"context"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/service"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

	passwordEncoder := component.NewPasswordEncoder()
	jwtService := service.NewJwtService(s.config.SecurityConfig, dataSource)
	userDetailDecoder := service.NewUserDetailDecoder(dataSource, jwtService)

	grpcServer := NewGrpcServer(s.config, dataSource, jwtService, userDetailDecoder, passwordEncoder).Start()
	httpServer := NewHttpServer(s.config, dataSource, jwtService, userDetailDecoder, passwordEncoder).Start()

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

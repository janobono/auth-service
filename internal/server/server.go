package server

import (
	"context"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	config      *config.ServerConfig
	initializer Initializer
}

func NewServer(config *config.ServerConfig, initializer Initializer) *Server {
	initSlog(config)
	if initializer == nil {
		initializer = NewInitializer()
	}
	return &Server{config, initializer}
}

func (s *Server) Start() {
	slog.Info("Starting server...")

	dataSource := db.NewDataSource(s.config.DbConfig)
	defer dataSource.Close()

	initDefaultCredentials(s.config, dataSource)

	repositories := s.initializer.Repositories(dataSource)

	utils := s.initializer.Utils(s.config)

	clients := s.initializer.Clients(s.config)
	defer clients.CaptchaClient.Close()

	services := s.initializer.Services(s.config, repositories, utils, clients)

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

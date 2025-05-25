package server

import (
	"errors"
	"github.com/janobono/auth-service/internal/component"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/service"
	"github.com/janobono/auth-service/pkg/security"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	config            *config.ServerConfig
	dataSource        *db.DataSource
	jwtService        service.JwtService
	userDetailDecoder security.UserDetailDecoder
	passwordEncoder   component.PasswordEncoder
}

func NewHttpServer(
	config *config.ServerConfig,
	dataSource *db.DataSource,
	jwtService service.JwtService,
	userDetailDecoder security.UserDetailDecoder,
	passwordEncoder component.PasswordEncoder,
) *HttpServer {
	return &HttpServer{config, dataSource, jwtService, userDetailDecoder, passwordEncoder}
}

func (s *HttpServer) Start() *http.Server {
	slog.Info("Starting http server...")

	// TODO implement

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	httpServer := &http.Server{
		Addr:    s.config.HTTPAddress,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to serve", "error", err)
			panic(err)
		}
	}()

	slog.Info("Http server started", "port", s.config.HTTPAddress)
	return httpServer
}

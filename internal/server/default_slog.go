package server

import (
	"log/slog"
	"os"

	"github.com/janobono/auth-service/internal/config"
)

func initSlog(config *config.ServerConfig) {
	var handler slog.Handler
	if config.Prod {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	slog.SetDefault(slog.New(handler))
}

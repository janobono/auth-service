package main

import (
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/server"
)

func main() {
	serverConfig := config.InitConfig()
	app := server.NewServer(serverConfig, nil)
	app.Start()
}

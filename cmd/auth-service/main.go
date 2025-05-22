package main

import (
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/server"
)

func main() {
	gRPCServerConfig := config.InitConfig()
	gRPCServer := server.New(gRPCServerConfig)
	gRPCServer.Start()
}

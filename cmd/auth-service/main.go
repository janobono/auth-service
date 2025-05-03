package main

import (
	"fmt"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
)

func main() {
	appConfig := config.InitConfig()

	pool := db.InitDb(appConfig.DbConfig)
	defer pool.Close()

	fmt.Println("auth-service")
}

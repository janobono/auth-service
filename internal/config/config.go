package config

import (
	"github.com/janobono/auth-service/pkg/util"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerConfig ServerConfig
	DbConfig     DbConfig
	AppConfig    AppConfig
}

type ServerConfig struct {
	Address string
}

type DbConfig struct {
	DBUrl      string
	DBUser     string
	DBPassword string
	DBMaxConns int
	DBMinConns int
}

type AppConfig struct {
	TokenIssuer    string
	TokenExpiresIn int
}

func InitConfig() *Config {
	godotenv.Load()

	return &Config{
		ServerConfig: ServerConfig{
			Address: getEnv("ADDRESS"),
		},
		DbConfig: DbConfig{
			DBUrl:      getEnv("DB_URL"),
			DBUser:     getEnv("DB_USER"),
			DBPassword: getEnv("DB_PASSWORD"),
			DBMaxConns: getEnvInt("DB_MAX_CONNS"),
			DBMinConns: getEnvInt("DB_MIN_CONNS"),
		},
		AppConfig: AppConfig{
			TokenIssuer:    getEnv("TOKEN_ISSUER"),
			TokenExpiresIn: getEnvInt("TOKEN_EXPIRES_IN"),
		},
	}
}

func getEnv(key string) string {
	result := ""
	if env, ok := os.LookupEnv(key); ok {
		result = env
	}

	if util.IsBlank(result) {
		log.Fatalf("configuration property %s not set", key)
	}
	return result
}

func getEnvInt(key string) int {
	s := getEnv(key)
	result, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("configuration property %s wrong format %v", key, err)
	}
	return result
}

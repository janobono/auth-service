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

func InitConfig() *Config {
	godotenv.Load()

	return &Config{
		ServerConfig: ServerConfig{
			Address: getEnv("ADDRESS", ":50052", true),
		},
		DbConfig: DbConfig{
			DBUrl:      getEnv("DB_URL", "localhost:5432/app", true),
			DBUser:     getEnv("DB_USER", "app", true),
			DBPassword: getEnv("DB_PASSWORD", "app", true),
			DBMaxConns: getEnvInt("DB_MAX_CONNS", "5", true),
			DBMinConns: getEnvInt("DB_MIN_CONNS", "2", true),
		},
	}
}

func getEnv(key, defaultValue string, required bool) string {
	result := defaultValue
	if env, ok := os.LookupEnv(key); ok {
		result = env
	}

	if required && util.IsBlank(result) {
		log.Fatalf("configuration property %s not set", key)
	}
	return result
}

func getEnvInt(key, defaultValue string, required bool) int {
	s := getEnv(key, defaultValue, required)
	result, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("configuration property %s wrong format %v", key, err)
	}
	return result
}

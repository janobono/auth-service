package config

import (
	"github.com/janobono/auth-service/pkg/util"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Prod           bool
	GRPCAddress    string
	HTTPAddress    string
	ContextPath    string
	DbConfig       *DbConfig
	MailConfig     *MailConfig
	SecurityConfig *SecurityConfig
	AppConfig      *AppConfig
}

type DbConfig struct {
	Url            string
	User           string
	Password       string
	MaxConnections int
	MinConnections int
	MigrationsUrl  string
}

type MailConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	AuthEnabled bool
	TlsEnabled  bool
}

type SecurityConfig struct {
	AuthorityAdmin           string
	AuthorityManager         string
	DefaultUsername          string
	DefaultPassword          string
	TokenIssuer              string
	AccessTokenExpiresIn     time.Duration
	AccessTokenJwkExpiresIn  time.Duration
	RefreshTokenExpiresIn    time.Duration
	RefreshTokenJwkExpiresIn time.Duration
	ContentTokenExpiresIn    time.Duration
	ContentTokenJwkExpiresIn time.Duration
}

type AppConfig struct {
	MailConfirmation bool
	ConfirmationUrl  string
}

func InitConfig() *ServerConfig {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Println("No .env.local file found")
	}

	return &ServerConfig{
		Prod:        getEnvBool("PROD"),
		GRPCAddress: getEnv("GRPC_ADDRESS"),
		HTTPAddress: getEnv("HTTP_ADDRESS"),
		ContextPath: getEnv("CONTEXT_PATH"),
		DbConfig: &DbConfig{
			Url:            getEnv("DB_URL"),
			User:           getEnv("DB_USER"),
			Password:       getEnv("DB_PASSWORD"),
			MaxConnections: getEnvInt("DB_MAX_CONNECTIONS"),
			MinConnections: getEnvInt("DB_MIN_CONNECTIONS"),
			MigrationsUrl:  getEnv("DB_MIGRATIONS_URL"),
		},
		MailConfig: &MailConfig{
			Host:        getEnv("MAIL_HOST"),
			Port:        getEnvInt("MAIL_PORT"),
			User:        getEnv("MAIL_USER"),
			Password:    getEnv("MAIL_PASSWORD"),
			AuthEnabled: getEnvBool("MAIL_AUTH_ENABLED"),
			TlsEnabled:  getEnvBool("MAIL_TLS_ENABLED"),
		},
		SecurityConfig: &SecurityConfig{
			AuthorityAdmin:           getEnv("SECURITY_AUTHORITY_ADMIN"),
			AuthorityManager:         getEnv("SECURITY_AUTHORITY_MANAGER"),
			DefaultUsername:          getEnv("SECURITY_DEFAULT_USERNAME"),
			DefaultPassword:          getEnv("SECURITY_DEFAULT_PASSWORD"),
			TokenIssuer:              getEnv("SECURITY_TOKEN_ISSUER"),
			AccessTokenExpiresIn:     time.Duration(getEnvInt("SECURITY_ACCESS_TOKEN_EXPIRES_IN")) * time.Minute,
			AccessTokenJwkExpiresIn:  time.Duration(getEnvInt("SECURITY_ACCESS_TOKEN_JWK_EXPIRES_IN")) * time.Minute,
			RefreshTokenExpiresIn:    time.Duration(getEnvInt("SECURITY_REFRESH_TOKEN_EXPIRES_IN")) * time.Minute,
			RefreshTokenJwkExpiresIn: time.Duration(getEnvInt("SECURITY_REFRESH_TOKEN_JWK_EXPIRES_IN")) * time.Minute,
			ContentTokenExpiresIn:    time.Duration(getEnvInt("SECURITY_CONTENT_TOKEN_EXPIRES_IN")) * time.Minute,
			ContentTokenJwkExpiresIn: time.Duration(getEnvInt("SECURITY_CONTENT_TOKEN_JWK_EXPIRES_IN")) * time.Minute,
		},
		AppConfig: &AppConfig{
			MailConfirmation: getEnvBool("APP_MAIL_CONFIRMATION"),
			ConfirmationUrl:  getEnv("APP_CONFIRMATION_URL"),
		},
	}
}

func getEnv(key string) string {
	result := ""
	if env, ok := os.LookupEnv(key); ok {
		result = env
	}

	if util.IsBlank(result) {
		log.Fatalf("Configuration property %s not set", key)
	}
	return result
}

func getEnvInt(key string) int {
	s := getEnv(key)
	result, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Configuration property %s wrong format %v", key, err)
	}
	return result
}

func getEnvBool(key string) bool {
	s := getEnv(key)
	result, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatalf("Configuration property %s wrong format %v", key, err)
	}
	return result
}

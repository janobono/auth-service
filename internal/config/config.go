package config

import (
	"log"
	"time"

	"github.com/janobono/go-util/common"
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
	CorsConfig     *CorsConfig
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
	ReadAuthorities          []string
	WriteAuthorities         []string
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

type CorsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	ExposedHeaders   []string
	MaxAge           time.Duration
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
		Prod:        common.EnvBool("PROD"),
		GRPCAddress: common.Env("GRPC_ADDRESS"),
		HTTPAddress: common.Env("HTTP_ADDRESS"),
		ContextPath: common.Env("CONTEXT_PATH"),
		DbConfig: &DbConfig{
			Url:            common.Env("DB_URL"),
			User:           common.Env("DB_USER"),
			Password:       common.Env("DB_PASSWORD"),
			MaxConnections: common.EnvInt("DB_MAX_CONNECTIONS"),
			MinConnections: common.EnvInt("DB_MIN_CONNECTIONS"),
			MigrationsUrl:  common.Env("DB_MIGRATIONS_URL"),
		},
		MailConfig: &MailConfig{
			Host:        common.Env("MAIL_HOST"),
			Port:        common.EnvInt("MAIL_PORT"),
			User:        common.Env("MAIL_USER"),
			Password:    common.Env("MAIL_PASSWORD"),
			AuthEnabled: common.EnvBool("MAIL_AUTH_ENABLED"),
			TlsEnabled:  common.EnvBool("MAIL_TLS_ENABLED"),
		},
		SecurityConfig: &SecurityConfig{
			ReadAuthorities:          common.EnvSlice("SECURITY_READ_AUTHORITIES"),
			WriteAuthorities:         common.EnvSlice("SECURITY_WRITE_AUTHORITIES"),
			DefaultUsername:          common.Env("SECURITY_DEFAULT_USERNAME"),
			DefaultPassword:          common.Env("SECURITY_DEFAULT_PASSWORD"),
			TokenIssuer:              common.Env("SECURITY_TOKEN_ISSUER"),
			AccessTokenExpiresIn:     time.Duration(common.EnvInt("SECURITY_ACCESS_TOKEN_EXPIRES_IN")) * time.Minute,
			AccessTokenJwkExpiresIn:  time.Duration(common.EnvInt("SECURITY_ACCESS_TOKEN_JWK_EXPIRES_IN")) * time.Minute,
			RefreshTokenExpiresIn:    time.Duration(common.EnvInt("SECURITY_REFRESH_TOKEN_EXPIRES_IN")) * time.Minute,
			RefreshTokenJwkExpiresIn: time.Duration(common.EnvInt("SECURITY_REFRESH_TOKEN_JWK_EXPIRES_IN")) * time.Minute,
			ContentTokenExpiresIn:    time.Duration(common.EnvInt("SECURITY_CONTENT_TOKEN_EXPIRES_IN")) * time.Minute,
			ContentTokenJwkExpiresIn: time.Duration(common.EnvInt("SECURITY_CONTENT_TOKEN_JWK_EXPIRES_IN")) * time.Minute,
		},
		CorsConfig: &CorsConfig{
			AllowedOrigins:   common.EnvSlice("CORS_ALLOWED_ORIGINS"),
			AllowedMethods:   common.EnvSlice("CORS_ALLOWED_METHODS"),
			AllowedHeaders:   common.EnvSlice("CORS_ALLOWED_HEADERS"),
			ExposedHeaders:   common.EnvSlice("CORS_EXPOSED_HEADERS"),
			AllowCredentials: common.EnvBool("CORS_ALLOW_CREDENTIALS"),
			MaxAge:           time.Duration(common.EnvInt("CORS_MAX_AGE")) * time.Hour,
		},
		AppConfig: &AppConfig{
			MailConfirmation: common.EnvBool("APP_MAIL_CONFIRMATION"),
			ConfirmationUrl:  common.Env("APP_CONFIRMATION_URL"),
		},
	}
}

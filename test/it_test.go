package test

import (
	"context"
	"fmt"
	"github.com/janobono/auth-service/generated/proto"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/db"
	"github.com/janobono/auth-service/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"net"
	"syscall"
	"testing"
	"time"
)

type testInitializer struct {
	initializer server.Initializer
}

func (ti *testInitializer) Repositories(dataSource *db.DataSource) *server.Repositories {
	return ti.initializer.Repositories(dataSource)
}

func (ti *testInitializer) Utils(serverConfig *config.ServerConfig) *server.Utils {
	return ti.initializer.Utils(serverConfig)
}

func (ti *testInitializer) Clients(serverConfig *config.ServerConfig) *server.Clients {
	return &server.Clients{
		CaptchaClient: &testCaptchaClient{},
		MailClient:    &testMailClient{},
	}
}

func (ti *testInitializer) Services(serverConfig *config.ServerConfig, repositories *server.Repositories, utils *server.Utils, clients *server.Clients) *server.Services {
	return ti.initializer.Services(serverConfig, repositories, utils, clients)
}

func TestIntegrationSomething(t *testing.T) {
	freePorts, err := getFreePorts(2)
	if err != nil {
		t.Fatalf("failed to get free ports: %v", err)
	}

	serverConfig := &config.ServerConfig{
		Prod:        false,
		GRPCAddress: (*freePorts)[0],
		HTTPAddress: (*freePorts)[1],
		ContextPath: "/api",
		DbConfig:    DbConfig,
		MailConfig: &config.MailConfig{
			Host:                       "",
			Port:                       0,
			User:                       "",
			Password:                   "",
			AuthEnabled:                false,
			TlsEnabled:                 false,
			MailTemplateUrl:            "",
			MailTemplateReloadInterval: time.Duration(0),
		},
		SecurityConfig: &config.SecurityConfig{
			ReadAuthorities:          []string{"customer", "manager"},
			WriteAuthorities:         []string{"admin"},
			DefaultUsername:          "simple@auth.org",
			DefaultPassword:          "$2a$10$gRKMsjTON2A4b5PDIgjej.EZPvzVaKRj52Mug/9bfQBzAYmVF0Cae",
			TokenIssuer:              "simple",
			AccessTokenExpiresIn:     time.Duration(30) * time.Minute,
			AccessTokenJwkExpiresIn:  time.Duration(720) * time.Minute,
			RefreshTokenExpiresIn:    time.Duration(10080) * time.Minute,
			RefreshTokenJwkExpiresIn: time.Duration(20160) * time.Minute,
			ContentTokenExpiresIn:    time.Duration(10080) * time.Minute,
			ContentTokenJwkExpiresIn: time.Duration(20160) * time.Minute,
		},
		CorsConfig: &config.CorsConfig{
			AllowedOrigins:   []string{"*"}, // Or restrict to specific domains
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposedHeaders:   []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
		AppConfig: &config.AppConfig{
			CaptchaServiceUrl:  "",
			MailConfirmation:   true,
			ConfirmationUrl:    "http://localhost:3000/confirm",
			PasswordCharacters: "abcdefghijklmnopqrstuvwxyz0123456789",
			PasswordLength:     8,
		},
	}

	s := server.NewServer(serverConfig, &testInitializer{server.NewInitializer()})
	go s.Start()
	time.Sleep(500 * time.Millisecond)

	conn, err := grpc.NewClient(
		"localhost"+serverConfig.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	authClient := proto.NewAuthClient(conn)
	result, err := authClient.SignIn(context.Background(), &proto.SignInData{
		Email:    "simple@auth.org",
		Password: "simple",
	})
	if err != nil {
		t.Fatalf("failed to sign in: %v", err)
	}
	t.Logf("sign in result: %v", result)

	userClient := proto.NewUserClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+result.AccessToken,
	))
	usersPage, err := userClient.SearchUsers(ctx, &proto.SearchCriteria{})
	if err != nil {
		t.Fatalf("failed to search users: %v", err)
	}
	t.Logf("search result: %v", usersPage)

	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	time.Sleep(1 * time.Second)
}

func createTestGrpcClientConn(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func getFreePorts(count int) (*[]string, error) {
	var ports []string
	for i := 0; i < count; i++ {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, err
		}
		defer l.Close()

		addr := l.Addr().(*net.TCPAddr)
		ports = append(ports, fmt.Sprintf(":%d", addr.Port))
	}
	return &ports, nil
}

package test

import (
	"context"
	"fmt"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/internal/config"
	"github.com/janobono/auth-service/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"net"
	"syscall"
	"testing"
	"time"
)

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
		MailConfig:  MailConfig,
		SecurityConfig: &config.SecurityConfig{
			AuthorityAdmin:           "admin",
			AuthorityManager:         "manager",
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
		AppConfig: &config.AppConfig{
			MailConfirmation: true,
			ConfirmationUrl:  "http://localhost:3000/confirm",
		},
	}

	s := server.NewServer(serverConfig)
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

	authClient := authgrpc.NewAuthClient(conn)
	result, err := authClient.SignIn(context.Background(), &authgrpc.SignInData{
		Email:    "simple@auth.org",
		Password: "simple",
	})
	if err != nil {
		t.Fatalf("failed to sign in: %v", err)
	}
	t.Logf("sign in result: %v", result)

	userClient := authgrpc.NewUserClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+result.AccessToken,
	))
	usersPage, err := userClient.SearchUsers(ctx, &authgrpc.SearchCriteria{})
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

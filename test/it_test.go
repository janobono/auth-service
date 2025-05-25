package test

import (
	"context"
	"fmt"
	"github.com/janobono/auth-service/gen/authgrpc"
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
	s := server.New(TestConfig)
	go s.Start()
	time.Sleep(500 * time.Millisecond)

	conn, err := grpc.NewClient(
		"localhost"+TestConfig.GRPCAddress,
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

func getFreePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	addr := l.Addr().(*net.TCPAddr)
	return fmt.Sprintf(":%d", addr.Port), nil
}

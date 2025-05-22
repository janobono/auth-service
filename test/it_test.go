package test

import (
	"context"
	"github.com/janobono/auth-service/api"
	"github.com/janobono/auth-service/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"syscall"
	"testing"
	"time"
)

func TestIntegrationSomething(t *testing.T) {
	s := server.New(TestConfig)
	go s.Start()
	time.Sleep(500 * time.Millisecond)

	conn, err := grpc.NewClient(
		"localhost"+TestConfig.ServerConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	authClient := api.NewAuthServiceClient(conn)
	result, err := authClient.SignIn(context.Background(), &api.SignInData{
		Email:    "simple@auth.org",
		Password: "simple",
	})
	if err != nil {
		t.Fatalf("failed to sign in: %v", err)
	}
	t.Logf("sign in result: %v", result)

	userClient := api.NewUserServiceClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+result.Value,
	))
	usersPage, err := userClient.SearchUsers(ctx, &api.SearchCriteriaData{})
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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/janobono/auth-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	serverAddr := flag.String("addr", "localhost:8080", "gRPC server address")
	email := flag.String("email", "", "User email")
	password := flag.String("password", "", "User password")
	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("email and password are required")
	}

	conn, err := grpc.NewClient(
		*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := api.NewAuthServiceClient(conn)

	// Make the SignIn call
	resp, err := client.SignIn(context.Background(), &api.SignInData{
		Email:    *email,
		Password: *password,
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Println("JWT token:", resp.GetValue())
}

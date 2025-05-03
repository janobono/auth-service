package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const bearerPrefix = "Bearer "

// TokenValidator interface that defines the ValidateToken method
type TokenValidator interface {
	ValidateToken(token string) bool
}

// TokenInterceptor is the gRPC interceptor that checks Bearer token validation
type TokenInterceptor struct {
	TokenValidator TokenValidator
}

// CheckToken checks the token for specific methods
func (tokenInterceptor *TokenInterceptor) CheckToken(secureMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if the method needs authentication
		if secureMethods[info.FullMethod] {
			// Extract the token from the metadata
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
			}

			// Retrieve the token
			authHeader := md["authorization"]
			if len(authHeader) == 0 || !strings.HasPrefix(authHeader[0], bearerPrefix) {
				return nil, status.Errorf(codes.Unauthenticated, "missing or invalid Bearer token")
			}

			// Here, you can validate the token
			token := authHeader[0][len(bearerPrefix):]
			if !tokenInterceptor.TokenValidator.ValidateToken(token) {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}
		}

		// Proceed with the handler if no error occurred
		return handler(ctx, req)
	}
}

package client

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log/slog"
	"time"

	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/janobono/auth-service/generated/proto"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
)

type CaptchaClient interface {
	Validate(ctx context.Context, data *proto.CaptchaData) (*wrapperspb.BoolValue, error)
	RawClient() proto.CaptchaClient
	Close()
}

type captchaClient struct {
	conn    *grpc.ClientConn
	client  proto.CaptchaClient
	breaker *gobreaker.CircuitBreaker
}

var _ CaptchaClient = (*captchaClient)(nil)

func NewCaptchaClient(url string) (CaptchaClient, error) {
	conn, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			grpcRetry.UnaryClientInterceptor(
				grpcRetry.WithMax(3),
				grpcRetry.WithBackoff(grpcRetry.BackoffExponential(100*time.Millisecond)),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "CaptchaService",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},
	})

	return &captchaClient{
		conn:    conn,
		client:  proto.NewCaptchaClient(conn),
		breaker: breaker,
	}, nil
}

func (cc *captchaClient) Validate(ctx context.Context, data *proto.CaptchaData) (*wrapperspb.BoolValue, error) {
	result, err := cc.breaker.Execute(func() (interface{}, error) {
		return cc.client.Validate(ctx, data)
	})
	if err != nil {
		return nil, err
	}
	return result.(*wrapperspb.BoolValue), nil
}

func (cc *captchaClient) RawClient() proto.CaptchaClient {
	return cc.client
}

func (cc *captchaClient) Close() {
	err := cc.conn.Close()
	if err != nil {
		slog.Error("Failed to close captcha service", "error", err)
	}
}

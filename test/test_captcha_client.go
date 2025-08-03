package test

import (
	"context"
	"github.com/janobono/auth-service/generated/proto"
	client2 "github.com/janobono/auth-service/internal/service/client"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type testCaptchaClient struct {
}

var _ client2.CaptchaClient = (*testCaptchaClient)(nil)

func (tcc *testCaptchaClient) Validate(ctx context.Context, data *proto.CaptchaData) (*wrapperspb.BoolValue, error) {
	return &wrapperspb.BoolValue{Value: true}, nil
}

func (tcc *testCaptchaClient) RawClient() proto.CaptchaClient {
	return nil
}

func (tcc *testCaptchaClient) Close() {
}

package service

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/janobono/auth-service/internal/component"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type captchaServer struct {
	authgrpc.UnimplementedCaptchaServer
	passwordEncoder *component.PasswordEncoder
}

func NewCaptchaServer(passwordEncoder *component.PasswordEncoder) authgrpc.CaptchaServer {
	return &captchaServer{
		passwordEncoder: passwordEncoder,
	}
}

func (cs *captchaServer) IsValid(ctx context.Context, captchaData *authgrpc.CaptchaData) (*wrapperspb.BoolValue, error) {
	err := cs.passwordEncoder.Compare(captchaData.Text, captchaData.Token)
	if err != nil {
		return &wrapperspb.BoolValue{Value: false}, nil
	}
	return &wrapperspb.BoolValue{Value: true}, nil
}

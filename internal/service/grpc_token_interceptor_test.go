package service

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGrpcUserDetail(t *testing.T) {
	ctx := context.Background()

	user := GetGrpcUserDetail(ctx)

	assert.Equal(t, user == nil, true)

	ctx = context.WithValue(ctx, userDetailKey, &authgrpc.UserDetail{})

	user = GetGrpcUserDetail(ctx)

	assert.Equal(t, user != nil, true)
}

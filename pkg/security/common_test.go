package security

import (
	"context"
	"github.com/janobono/auth-service/gen/authgrpc"
	"gotest.tools/v3/assert"
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

func TestHasAnyAuthority(t *testing.T) {
	type testCase struct {
		methodAuthorities []string
		userAuthorities   []string
		expected          bool
	}

	t.Run("HasAnyAuthority", func(t *testing.T) {
		tests := []testCase{
			{methodAuthorities: []string{}, userAuthorities: []string{}, expected: false},
			{methodAuthorities: []string{"test"}, userAuthorities: []string{}, expected: false},
			{methodAuthorities: []string{"test"}, userAuthorities: []string{"test"}, expected: true},
			{methodAuthorities: []string{"test1", "test2"}, userAuthorities: []string{"test"}, expected: false},
			{methodAuthorities: []string{"test1", "test2"}, userAuthorities: []string{"test2"}, expected: true},
		}

		for _, test := range tests {
			actual := HasAnyAuthority(&test.methodAuthorities, &test.userAuthorities)
			assert.Equal(t, actual, test.expected)
		}
	})
}

func TestFindGrpcSecuredMethod(t *testing.T) {
	type testCase struct {
		securedMethods []GrpcSecuredMethod
		methodName     string
		found          bool
	}

	t.Run("findGrpcSecuredMethod", func(t *testing.T) {
		tests := []testCase{
			{securedMethods: []GrpcSecuredMethod{}, methodName: "", found: false},
			{securedMethods: []GrpcSecuredMethod{}, methodName: authgrpc.Auth_GetUser_FullMethodName, found: false},
			{securedMethods: []GrpcSecuredMethod{
				{Method: authgrpc.Auth_GetUser_FullMethodName, Authorities: []string{"test"}},
			}, methodName: "", found: false},
			{securedMethods: []GrpcSecuredMethod{
				{Method: authgrpc.Auth_GetUser_FullMethodName, Authorities: []string{"test"}},
			}, methodName: authgrpc.Auth_GetUser_FullMethodName, found: true},
		}

		for _, test := range tests {
			actual := findGrpcSecuredMethod(&test.securedMethods, test.methodName)
			assert.Equal(t, actual != nil, test.found)
		}
	})
}

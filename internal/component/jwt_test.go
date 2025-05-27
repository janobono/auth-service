package component

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJwt(t *testing.T) {
	jwtToken := initJwtToken(t, 60)

	token, err := jwtToken.GenerateToken(jwt.MapClaims{"data": "test data"})
	if err != nil {
		t.Fatalf("Error generating token: %s", err)
	}

	t.Logf("Generated Token: %s", token)

	accessToken, err := jwtToken.GenerateAccessToken("id", []string{"role1", "role2"})
	if err != nil {
		t.Fatalf("Error generating accessToken: %s", err)
	}

	t.Logf("Generated accessToken: %s", accessToken)

	parsedToken, err := jwtToken.ParseToken(token)
	if err != nil {
		t.Fatalf("Error parsing token: %s", err)
	}
	assert.Equal(t, (*parsedToken)["data"], "test data")

	id, authorities, err := jwtToken.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("Error parsing accessToken: %s", err)
	}
	assert.Equal(t, id, "id")
	assert.Equal(t, (*authorities)[0], "role1")
}

func TestJwtExpired(t *testing.T) {
	jwtToken := initJwtToken(t, 0)

	token, err := jwtToken.GenerateToken(jwt.MapClaims{"data": "test data"})
	if err != nil {
		t.Fatalf("Error generating token: %s", err)
	}

	t.Logf("Generated Token: %s", token)

	time.Sleep(100 * time.Millisecond)

	_, err = jwtToken.ParseToken(token)

	assert.Error(t, err, "error parsing token: token has invalid claims: token is expired")
}

func TestJwtSignature(t *testing.T) {
	jwtToken1 := initJwtToken(t, 60)
	jwtToken2 := initJwtToken(t, 60)

	token, err := jwtToken1.GenerateToken(jwt.MapClaims{"data": "test data"})
	if err != nil {
		t.Fatalf("Error generating token: %s", err)
	}

	t.Logf("Generated Token: %s", token)

	_, err = jwtToken2.ParseToken(token)

	assert.Error(t, err, "error parsing token: token signature is invalid: crypto/rsa: verification error")
}

func initJwtToken(t *testing.T, expiration int64) *JwtToken {
	algorithm := jwt.SigningMethodRS256
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Error generating private key: %s", err)
	}
	publicKey := &privateKey.PublicKey

	jwtToken := NewJwtToken(
		algorithm,
		privateKey,
		publicKey,
		"kid",
		"test",
		time.Duration(expiration)*time.Second,
		time.Now(),
		func(kid string) (interface{}, error) {
			return nil, fmt.Errorf("not implemented")
		},
	)

	return jwtToken
}

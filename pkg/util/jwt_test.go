package util

import (
	"gotest.tools/v3/assert"
	"testing"
	"time"
)

func TestJwt(t *testing.T) {
	jwtToken, err := NewJwtToken(JwtConfigProperties{
		Issuer:     "test",
		Expiration: 60,
	})

	if err != nil {
		t.Fatalf("Error creating JWT token: %s", err)
	}

	jwtContent := JwtContent{
		ID:          123,
		Authorities: []string{"admin"},
	}

	token, err := jwtToken.GenerateToken(jwtContent, time.Now().Unix())
	if err != nil {
		t.Fatalf("Error generating token: %s", err)
	}

	t.Logf("Generated Token: %s", token)

	parsedContent, err := jwtToken.ParseToken(token)
	if err != nil {
		t.Fatalf("Error parsing token: %s", err)
	}

	assert.Equal(t, jwtContent.ID, parsedContent.ID)
	assert.Equal(t, len(jwtContent.Authorities), len(parsedContent.Authorities))
	assert.Equal(t, jwtContent.Authorities[0], parsedContent.Authorities[0])
}

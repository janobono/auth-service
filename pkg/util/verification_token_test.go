package util

import (
	"gotest.tools/v3/assert"
	"testing"
	"time"
)

func TestVerificationToken(t *testing.T) {
	verificationToken, err := NewVerificationToken(VerificationConfigProperties{
		Issuer: "test",
	})

	if err != nil {
		t.Fatalf("Error creating verification token: %s", err)
	}

	verificationContent := map[string]string{
		"test": "12345",
	}

	token, err := verificationToken.GenerateToken(verificationContent, time.Now().Unix(), time.Now().Unix()+60)
	if err != nil {
		t.Fatalf("Error generating token: %s", err)
	}

	t.Logf("Generated Token: %s", token)

	parsedContent, err := verificationToken.ParseToken(token)
	if err != nil {
		t.Fatalf("Error parsing token: %s", err)
	}

	assert.Equal(t, verificationContent["test"], parsedContent["test"])
}

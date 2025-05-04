package util

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type VerificationConfigProperties struct {
	Issuer string
}

type VerificationToken struct {
	Algorithm  *jwt.SigningMethodRSA
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Expiration int64
	Issuer     string
}

func NewVerificationToken(verificationConfigProperties VerificationConfigProperties) (*VerificationToken, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	publicKey := &privateKey.PublicKey

	algorithm := jwt.SigningMethodRS256
	return &VerificationToken{
		Algorithm:  algorithm,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Issuer:     verificationConfigProperties.Issuer,
	}, nil
}

func (vt *VerificationToken) GenerateToken(data map[string]string, issuedAt, expiresAt int64) (string, error) {
	claims := jwt.MapClaims{
		"iss": vt.Issuer,
		"iat": issuedAt,
		"exp": expiresAt,
		"sub": "verification",
	}

	for key, value := range data {
		claims[key] = value
	}

	token := jwt.NewWithClaims(vt.Algorithm, claims)
	signedToken, err := token.SignedString(vt.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (vt *VerificationToken) ParseToken(tokenString string) (map[string]string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != vt.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return vt.PublicKey, nil
	})

	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return map[string]string{}, fmt.Errorf("invalid token")
	}

	result := make(map[string]string)
	for key, value := range claims {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result, nil
}

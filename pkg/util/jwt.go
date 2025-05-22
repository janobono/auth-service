package util

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type JwtConfigProperties struct {
	Expiration int64
	Issuer     string
}

type JwtContent struct {
	ID          int64
	Authorities []string
}

type JwtToken struct {
	Algorithm  *jwt.SigningMethodRSA
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Expiration int64
	Issuer     string
}

func NewJwtToken(jwtConfigProperties *JwtConfigProperties) (*JwtToken, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	publicKey := &privateKey.PublicKey

	algorithm := jwt.SigningMethodRS256
	return &JwtToken{
		Algorithm:  algorithm,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Expiration: jwtConfigProperties.Expiration,
		Issuer:     jwtConfigProperties.Issuer,
	}, nil
}

func (t *JwtToken) GenerateToken(jwtContent *JwtContent, issuedAt int64) (string, error) {
	claims := jwt.MapClaims{
		"iss": t.Issuer,
		"iat": issuedAt,
		"exp": issuedAt + t.Expiration,
		"sub": fmt.Sprintf("%d", jwtContent.ID),
	}

	claims["aud"] = jwtContent.Authorities

	token := jwt.NewWithClaims(t.Algorithm, claims)
	signedToken, err := token.SignedString(t.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (t *JwtToken) ParseToken(tokenString string) (*JwtContent, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != t.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.PublicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	id := int64(0)
	if sub, ok := claims["sub"].(string); ok {
		fmt.Sscanf(sub, "%d", &id)
	}

	var authorities []string
	if aud, ok := claims["aud"].([]interface{}); ok {
		for _, a := range aud {
			authorities = append(authorities, a.(string))
		}
	}

	return &JwtContent{
		ID:          id,
		Authorities: authorities,
	}, nil
}

package repository

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/janobono/auth-service/generated/sqlc"
	"time"
)

type AddAttributeData struct {
	Key      string
	Required bool
	Hidden   bool
}

type AddAuthorityData struct {
	Authority string
}

type AddJwkData struct {
	Use        string
	Expiration time.Duration
}

type AddUserData struct {
	Email     string
	Password  string
	Confirmed bool
	Enabled   bool
}

type Attribute struct {
	ID       pgtype.UUID
	Key      string
	Required bool
	Hidden   bool
}

type Authority struct {
	ID        pgtype.UUID
	Authority string
}

type Jwk struct {
	ID         pgtype.UUID
	Kty        string
	Use        string
	Alg        string
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
	Active     bool
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type SearchUsersCriteria struct {
	SearchField   string
	Email         string
	AttributeKeys []string
}

type SetUserAttributesData struct {
	UserID     pgtype.UUID
	Attributes []*UserAttribute
}

type SetUserAuthoritiesData struct {
	UserID      pgtype.UUID
	Authorities []*Authority
}

type User struct {
	ID        pgtype.UUID
	CreatedAt time.Time
	Email     string
	Password  string
	Confirmed bool
	Enabled   bool
}

type UserAttribute struct {
	Attribute *Attribute
	Value     string
}

func encodePrivateKey(privateKey *rsa.PrivateKey) []byte {
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}
	return pem.EncodeToMemory(block)
}

func encodePublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	}
	return pem.EncodeToMemory(block), nil
}

func parsePrivate(jwk *sqlc.Jwk) (*rsa.PrivateKey, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(jwk.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	return privateKey, nil
}

func parsePublicKey(jwk *sqlc.Jwk) (*rsa.PublicKey, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(jwk.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}
	return publicKey, nil
}

func toAttribute(attribute *sqlc.Attribute) *Attribute {
	return &Attribute{
		ID:       attribute.ID,
		Key:      attribute.Key,
		Required: attribute.Required,
		Hidden:   attribute.Hidden,
	}
}

func toAuthority(authority *sqlc.Authority) *Authority {
	return &Authority{
		ID:        authority.ID,
		Authority: authority.Authority,
	}
}

func toJwk(jwk *sqlc.Jwk) (*Jwk, error) {
	privateKey, err := parsePrivate(jwk)

	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKey(jwk)

	if err != nil {
		return nil, err
	}

	return &Jwk{
		ID:         jwk.ID,
		Kty:        jwk.Kty,
		Use:        jwk.Use,
		Alg:        jwk.Alg,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Active:     jwk.Active,
		CreatedAt:  jwk.CreatedAt.Time,
		ExpiresAt:  jwk.ExpiresAt.Time,
	}, nil
}

func toUser(user *sqlc.User) *User {
	return &User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		Email:     user.Email,
		Password:  user.Password,
		Confirmed: user.Confirmed,
		Enabled:   user.Enabled,
	}
}

func toUserAttribute(attribute *sqlc.GetUserAttributesRow) *UserAttribute {
	return &UserAttribute{
		Attribute: &Attribute{
			ID:       attribute.ID,
			Key:      attribute.Key,
			Required: attribute.Required,
			Hidden:   attribute.Hidden,
		},
		Value: func() string {
			if attribute.Value.Valid {
				return attribute.Value.String
			}
			return ""
		}(),
	}
}

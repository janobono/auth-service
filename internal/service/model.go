package service

import (
	"crypto/rsa"
	"time"
)

const (
	ErrNotFound         = "NOT_FOUND"
	ErrPermissionDenied = "PERMISSION_DENIED"
	ErrInternalError    = "INTERNAL_ERROR"
	InvalidArgument     = "INVALID_ARGUMENT"
	InvalidCredentials  = "INVALID_CREDENTIALS"
	UserDisabled        = "USER_DISABLED"
	UserNotConfirmed    = "USER_NOT_CONFIRMED"
)

type Attribute struct {
	Id       string
	Key      string
	Name     string
	Required bool
	Hidden   bool
}

type Authority struct {
	Id        string
	Authority string
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
}

type Jwk struct {
	Id         string
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

type SignIn struct {
	Email    string
	Password string
}

type UserDetail struct {
	Id          string
	CreatedAt   time.Time
	Email       string
	Password    string
	Confirmed   bool
	Enabled     bool
	Attributes  []*UserAttribute
	Authorities []*Authority
}

type UserAttribute struct {
	Attribute *Attribute
	Value     string
}

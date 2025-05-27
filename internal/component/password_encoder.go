package component

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type PasswordEncoder struct {
}

func NewPasswordEncoder() *PasswordEncoder {
	return &PasswordEncoder{}
}

func (p *PasswordEncoder) Encode(password string) (string, error) {
	encodedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("unable to encrypt password")
	}
	return string(encodedPassword), nil
}

func (p *PasswordEncoder) Compare(password, encodedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(password))
}

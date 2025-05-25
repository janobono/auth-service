package component

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type PasswordEncoder interface {
	Encode(password string) (string, error)
	Compare(password, encodedPassword string) error
}

type passwordEncoder struct {
}

func NewPasswordEncoder() PasswordEncoder {
	return &passwordEncoder{}
}

func (p *passwordEncoder) Encode(password string) (string, error) {
	encodedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("unable to encrypt password")
	}
	return string(encodedPassword), nil
}

func (p *passwordEncoder) Compare(password, encodedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(password))
}

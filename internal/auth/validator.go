package auth

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type CredentialsValidator struct{}

func NewCredentialsValidator() *CredentialsValidator {
	return &CredentialsValidator{}
}

func (v *CredentialsValidator) NormalizeAndValidate(
	request *LoginRequest,
) error {
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	if request.Email == "" || request.Password == "" {
		return ErrInvalidCredentials
	}

	return nil
}

func (v *CredentialsValidator) VerifyPassword(
	passwordHash string,
	password string,
) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))

	switch {
	case err == nil:
		return nil

	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return ErrInvalidCredentials

	default:
		return fmt.Errorf("compare password hash: %w", err)
	}
}

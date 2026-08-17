package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"mmktestbasisByDGanichev/internal/database"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrEmailTooLong     = errors.New("email is too long")
	ErrNameRequired     = errors.New("name is required")
	ErrNameTooLong      = errors.New("name is too long")
	ErrPasswordTooShort = errors.New("password must contain at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 72 bytes")
)

type Service struct {
	db    database.Database
	users userCreator
}

type userCreator interface {
	Create(ctx context.Context, db database.DBTX, user User) (User, error)
}

func NewService(db database.Database, users userCreator) *Service {
	return &Service{
		db:    db,
		users: users,
	}
}

func (s *Service) Register(
	ctx context.Context,
	request RegisterRequest,
) (Response, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	name := strings.TrimSpace(request.Name)

	if err := validateRegistration(email, request.Password, name); err != nil {
		return Response{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return Response{}, fmt.Errorf("hash password: %w", err)
	}

	newUser := User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Name:         name,
		CreatedAt:    time.Now().UTC(),
	}

	var created User

	err = s.db.WithinTransaction(ctx, func(tx database.DBTX) error {
		var err error

		created, err = s.users.Create(ctx, tx, newUser)

		return err
	})
	if err != nil {
		return Response{}, err
	}

	return responseFromUser(created), nil
}

func validateRegistration(
	email string,
	password string,
	name string,
) error {
	switch {
	case email == "":
		return ErrEmailRequired

	case len(email) > 255:
		return ErrEmailTooLong

	case !isValidEmail(email):
		return ErrInvalidEmail

	case name == "":
		return ErrNameRequired

	case utf8.RuneCountInString(name) > 255:
		return ErrNameTooLong

	case utf8.RuneCountInString(password) < 8:
		return ErrPasswordTooShort

	case len([]byte(password)) > 72:
		// Ограничение bcrypt.
		return ErrPasswordTooLong

	default:
		return nil
	}
}

func isValidEmail(value string) bool {
	address, err := mail.ParseAddress(value)

	return err == nil && address.Address == value
}

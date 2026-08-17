package auth

import (
	"context"
	"errors"
	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/user"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type userFinder interface {
	FindByEmail(ctx context.Context, db database.DBTX, email string) (user.User, error)
}

type tokenIssuer interface {
	Issue(userID int64) (string, error)
}

type credentialsValidator interface {
	NormalizeAndValidate(request *LoginRequest) error

	VerifyPassword(passwordHash string, password string) error
}

type Service struct {
	db        database.DBTX
	users     userFinder
	tokens    tokenIssuer
	validator credentialsValidator
	tokenTTL  time.Duration
}

func NewService(
	db database.DBTX,
	users userFinder,
	tokens tokenIssuer,
	validator credentialsValidator,
	tokenTTL time.Duration,
) *Service {
	return &Service{
		db:        db,
		users:     users,
		tokens:    tokens,
		validator: validator,
		tokenTTL:  tokenTTL,
	}
}

func (s *Service) Login(
	ctx context.Context,
	request LoginRequest,
) (LoginResponse, error) {
	if err := s.validator.NormalizeAndValidate(
		&request,
	); err != nil {
		return LoginResponse{}, err
	}

	foundUser, err := s.users.FindByEmail(ctx, s.db, request.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return LoginResponse{}, ErrInvalidCredentials
		}

		return LoginResponse{}, err
	}

	if err := s.validator.VerifyPassword(
		foundUser.PasswordHash,
		request.Password,
	); err != nil {
		return LoginResponse{}, err
	}

	accessToken, err := s.tokens.Issue(foundUser.ID)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokenTTL.Seconds()),
	}, nil
}

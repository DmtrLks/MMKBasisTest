package auth

import (
	"errors"
	"fmt"
	"mmktestbasisByDGanichev/internal/config"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(cfg config.JWTConfig) *TokenManager {
	return &TokenManager{
		secret: []byte(cfg.Secret),
		issuer: cfg.Issuer,
		ttl:    cfg.TTL,
		now:    time.Now,
	}
}

func (m *TokenManager) Issue(userID int64) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("issue token: invalid user ID")
	}

	now := m.now().UTC()

	claims := jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}

func (m *TokenManager) Parse(tokenString string) (int64, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}

	return userID, nil
}

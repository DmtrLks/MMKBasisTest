package middleware

import (
	"context"
	"net/http"
	"strings"

	"mmktestbasisByDGanichev/internal/httpg"
)

type tokenParser interface {
	Parse(token string) (int64, error)
}

type Auth struct {
	tokens tokenParser
}

func NewAuth(tokens tokenParser) *Auth {
	return &Auth{
		tokens: tokens,
	}
}

func (m *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				httpg.WriteError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"valid bearer token is required",
				)

				return
			}

			userID, err := m.tokens.Parse(token)
			if err != nil {
				httpg.WriteError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"valid bearer token is required",
				)

				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey{}, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

func bearerToken(r *http.Request) (string, bool) {
	parts := strings.Fields(r.Header.Get("Authorization"))

	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	if parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

type userIDContextKey struct{}

func UserIDFromContext(
	ctx context.Context,
) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(int64)

	return userID, ok && userID > 0
}

func RequireUserID(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		httpg.WriteUnauthorized(w)

		return 0, false
	}

	return userID, true
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterRefillsTokens(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	limiter := NewRateLimiter(2, 2)
	limiter.now = func() time.Time { return now }

	if !limiter.allow("client") {
		t.Fatal("first request must be allowed")
	}
	if !limiter.allow("client") {
		t.Fatal("second request within burst must be allowed")
	}
	if limiter.allow("client") {
		t.Fatal("request above burst must be rejected")
	}

	now = now.Add(500 * time.Millisecond)
	if !limiter.allow("client") {
		t.Fatal("one token must be restored after half a second at 2 RPS")
	}
	if limiter.allow("client") {
		t.Fatal("restored token must be consumed")
	}
}

func TestRateLimiterSeparatesClients(t *testing.T) {
	limiter := NewRateLimiter(1, 1)

	if !limiter.allow("first") {
		t.Fatal("first client must be allowed")
	}
	if limiter.allow("first") {
		t.Fatal("first client must be throttled")
	}
	if !limiter.allow("second") {
		t.Fatal("second client must have its own bucket")
	}
}

func TestRateLimiterReturnsTooManyRequests(t *testing.T) {
	limiter := NewRateLimiter(1, 1)
	handler := limiter.Limit(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	first.RemoteAddr = "192.0.2.1:1234"
	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)

	if firstResponse.Code != http.StatusNoContent {
		t.Fatalf("first response status = %d, want %d", firstResponse.Code, http.StatusNoContent)
	}

	second := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	second.RemoteAddr = "192.0.2.1:5678"
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)

	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"second response status = %d, want %d",
			secondResponse.Code,
			http.StatusTooManyRequests,
		)
	}
	if secondResponse.Header().Get("Retry-After") != "1" {
		t.Fatal("Retry-After header must be set")
	}
}

func TestRateLimiterDoesNotLimitHealthCheck(t *testing.T) {
	limiter := NewRateLimiter(1, 1)
	handler := limiter.Limit(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for range 3 {
		request := httptest.NewRequest(http.MethodGet, "/health", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("health response status = %d, want %d", response.Code, http.StatusNoContent)
		}
	}
}

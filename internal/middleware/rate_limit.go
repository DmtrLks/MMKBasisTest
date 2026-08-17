package middleware

import (
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"mmktestbasisByDGanichev/internal/httpg"
)

const (
	clientTTL       = 10 * time.Minute
	cleanupInterval = time.Minute
)

type clientBucket struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

type RateLimiter struct {
	mu          sync.Mutex
	clients     map[string]*clientBucket
	rate        float64
	burst       float64
	now         func() time.Time
	lastCleanup time.Time
}

func NewRateLimiter(
	requestsPerSecond int,
	burst int,
) *RateLimiter {
	now := time.Now()

	return &RateLimiter{
		clients:     make(map[string]*clientBucket),
		rate:        float64(requestsPerSecond),
		burst:       float64(burst),
		now:         time.Now,
		lastCleanup: now,
	}
}

func (l *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health checks must continue working even when API clients are throttled.
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		if !l.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "1")
			httpg.WriteError(
				w,
				http.StatusTooManyRequests,
				"rate_limit_exceeded",
				"too many requests",
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) allow(client string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if now.Sub(l.lastCleanup) >= cleanupInterval {
		l.cleanup(now)
		l.lastCleanup = now
	}

	bucket, exists := l.clients[client]
	if !exists {
		l.clients[client] = &clientBucket{
			tokens:     l.burst - 1,
			lastRefill: now,
			lastSeen:   now,
		}
		return true
	}

	elapsed := now.Sub(bucket.lastRefill).Seconds()
	if elapsed > 0 {
		bucket.tokens = math.Min(l.burst, bucket.tokens+elapsed*l.rate)
		bucket.lastRefill = now
	}

	bucket.lastSeen = now
	if bucket.tokens < 1 {
		return false
	}

	bucket.tokens--
	return true
}

func (l *RateLimiter) cleanup(now time.Time) {
	for client, bucket := range l.clients {
		if now.Sub(bucket.lastSeen) >= clientTTL {
			delete(l.clients, client)
		}
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

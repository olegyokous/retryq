// Package ratelimit provides token-bucket rate limiting for the retry queue.
package ratelimit

import (
	"net/http"
	"time"
)

// StatusTooManyRequests is returned when the rate limit is exceeded.
const StatusTooManyRequests = http.StatusTooManyRequests

// RetryAfterHeader is the header set when a request is rate-limited.
const RetryAfterHeader = "Retry-After"

// Middleware wraps an http.Handler and enforces rate limiting.
// Requests that exceed the configured rate receive a 429 response.
type Middleware struct {
	limiter  *Limiter
	retryAfter time.Duration
}

// NewMiddleware creates a Middleware that gates requests through the given Limiter.
// retryAfter controls the value of the Retry-After header on 429 responses.
func NewMiddleware(l *Limiter, retryAfter time.Duration) *Middleware {
	if retryAfter <= 0 {
		retryAfter = time.Second
	}
	return &Middleware{limiter: l, retryAfter: retryAfter}
}

// Handler returns an http.Handler that enforces rate limiting before delegating
// to next.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := m.limiter.Allow(); err != nil {
			retrySeconds := int(m.retryAfter.Seconds())
			w.Header().Set(RetryAfterHeader, itoa(retrySeconds))
			http.Error(w, "rate limit exceeded", StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// itoa converts an int to its decimal string representation without importing
// strconv at the cost of a tiny allocation for the common small-int case.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

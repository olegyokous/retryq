// Package circuitbreaker provides a circuit breaker implementation
// for protecting downstream services from cascading failures.
package circuitbreaker

import (
	"net/http"
	"strconv"
)

const (
	// HeaderCircuitState is the response header indicating the circuit state.
	HeaderCircuitState = "X-Circuit-State"
	// HeaderRetryAfter is the response header indicating when to retry.
	HeaderRetryAfter = "Retry-After"
)

// Middleware returns an http.Handler that rejects requests when the circuit
// breaker is open, and records success/failure outcomes after each request.
func NewMiddleware(cb *CircuitBreaker, next http.Handler) http.Handler {
	if cb == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := cb.Allow(); err != nil {
			cooldown := int(cb.opts.CooldownPeriod.Seconds())
			w.Header().Set(HeaderCircuitState, "open")
			w.Header().Set(HeaderRetryAfter, strconv.Itoa(cooldown))
			http.Error(w, "circuit breaker open: "+err.Error(), http.StatusServiceUnavailable)
			return
		}

		rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)

		if rw.code >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}

		w.Header().Set(HeaderCircuitState, cb.State())
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

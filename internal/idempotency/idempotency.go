// Package idempotency provides a middleware that rejects duplicate requests
// based on a caller-supplied idempotency key header. Keys are tracked in a
// bounded, TTL-aware in-memory store backed by the dedupe package.
package idempotency

import (
	"net/http"
	"time"

	"github.com/andrebq/retryq/internal/dedupe"
)

const (
	// DefaultHeader is the HTTP header inspected for the idempotency key.
	DefaultHeader = "Idempotency-Key"

	// DefaultTTL is how long a key is remembered after first use.
	DefaultTTL = 24 * time.Hour

	// DefaultMaxKeys is the maximum number of keys kept in memory.
	DefaultMaxKeys = 10_000
)

// Options configures the idempotency middleware.
type Options struct {
	// Header is the request header that carries the idempotency key.
	// Defaults to DefaultHeader.
	Header string

	// TTL controls how long a key is considered active.
	// Defaults to DefaultTTL.
	TTL time.Duration

	// MaxKeys is the maximum number of live keys tracked.
	// Defaults to DefaultMaxKeys.
	MaxKeys int
}

// DefaultOptions returns an Options populated with package-level defaults.
func DefaultOptions() Options {
	return Options{
		Header:  DefaultHeader,
		TTL:     DefaultTTL,
		MaxKeys: DefaultMaxKeys,
	}
}

// NewMiddleware wraps next and rejects requests whose idempotency key has
// already been seen within the configured TTL window. A nil opts pointer
// causes the middleware to use DefaultOptions.
func NewMiddleware(next http.Handler, opts *Options) http.Handler {
	o := DefaultOptions()
	if opts != nil {
		o = *opts
	}
	if o.Header == "" {
		o.Header = DefaultHeader
	}

	dd := dedupe.New(dedupe.Options{
		TTL:     o.TTL,
		MaxSize: o.MaxKeys,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(o.Header)
		if key == "" {
			// No key supplied — pass through without tracking.
			next.ServeHTTP(w, r)
			return
		}

		if dd.IsDuplicate(key) {
			http.Error(w, "duplicate request", http.StatusConflict)
			return
		}

		next.ServeHTTP(w, r)
	})
}

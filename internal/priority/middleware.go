// Package priority defines priority levels for queued items and provides
// HTTP middleware that reads a priority hint from the incoming request.
package priority

import (
	"net/http"
)

const (
	// HeaderPriority is the HTTP header used to signal the desired priority.
	HeaderPriority = "X-Retry-Priority"

	// contextKey is the unexported type used to store Priority in a context.
	contextKey struct{}
)

// Middleware extracts a Priority from the X-Retry-Priority request header and
// stores it in the request context so downstream handlers can read it without
// re-parsing the header themselves.
//
// If the header is absent or contains an unrecognised value the request is
// passed through with the default (Normal) priority; no error is returned to
// the caller.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := Normal
		if raw := r.Header.Get(HeaderPriority); raw != "" {
			if parsed, err := Parse(raw); err == nil {
				p = parsed
			}
		}
		ctx := r.Context()
		ctx = contextWithPriority(ctx, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FromContext returns the Priority stored in ctx, falling back to Normal when
// no value has been set.
func FromContext(r *http.Request) Priority {
	if p, ok := r.Context().Value(contextKey{}).(Priority); ok {
		return p
	}
	return Normal
}

// contextWithPriority returns a copy of ctx carrying p.
func contextWithPriority(ctx interface{ Value(any) any }, p Priority) interface{ Value(any) any } {
	// We need a real context.Context; import it via the standard library.
	return priorityCtx{parent: ctx.(interface {
		Value(any) any
	}), val: p}
}

// Package telemetry provides structured request tracing for retryq,
// attaching a unique trace ID to every enqueued item and propagating it
// through retries so operators can correlate log lines end-to-end.
package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

type contextKey struct{}

// TraceID is a 16-byte hex-encoded random identifier.
type TraceID string

// New generates a cryptographically random TraceID.
func New() (TraceID, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return TraceID(hex.EncodeToString(b)), nil
}

// MustNew generates a TraceID and panics on error (suitable for tests).
func MustNew() TraceID {
	id, err := New()
	if err != nil {
		panic(err)
	}
	return id
}

// WithContext stores a TraceID in ctx.
func WithContext(ctx context.Context, id TraceID) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// FromContext retrieves the TraceID from ctx.
// Returns an empty TraceID and false if none is present.
func FromContext(ctx context.Context) (TraceID, bool) {
	id, ok := ctx.Value(contextKey{}).(TraceID)
	return id, ok && id != ""
}

// Header is the canonical HTTP header used to propagate trace IDs.
const Header = "X-Retryq-Trace-Id"

// Middleware extracts an incoming trace ID from the request header, or
// generates a fresh one, then stores it in the request context and echoes
// it back in the response header.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := TraceID(r.Header.Get(Header))
		if id == "" {
			generated, err := New()
			if err != nil {
				slog.Error("telemetry: failed to generate trace id", "err", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			id = generated
		}
		w.Header().Set(Header, string(id))
		next.ServeHTTP(w, r.WithContext(WithContext(r.Context(), id)))
	})
}

// LogAttr returns an [slog.Attr] for the trace ID, suitable for structured
// log lines.
func LogAttr(id TraceID) slog.Attr {
	return slog.String("trace_id", string(id))
}

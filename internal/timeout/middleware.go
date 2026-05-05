// Package timeout provides per-request deadline enforcement for HTTP handlers.
// It wraps an http.Handler and cancels the request context after a configurable
// duration, returning 503 Service Unavailable to the caller.
package timeout

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Options controls the behaviour of the timeout middleware.
type Options struct {
	// Timeout is the maximum duration allowed for a single request.
	// Defaults to 30 seconds when zero.
	Timeout time.Duration

	// Logger receives a warning whenever a request is cancelled due to timeout.
	// Falls back to slog.Default() when nil.
	Logger *slog.Logger
}

// DefaultOptions returns a sensible Options value ready for production use.
func DefaultOptions() Options {
	return Options{
		Timeout: 30 * time.Second,
		Logger:  slog.Default(),
	}
}

// NewMiddleware returns an http.Handler that enforces a per-request deadline.
// Requests that exceed opts.Timeout receive a 503 response; the wrapped handler
// is still given a cancelled context so it can abort in-flight work.
func NewMiddleware(next http.Handler, opts Options) http.Handler {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultOptions().Timeout
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), opts.Timeout)
		defer cancel()

		// done receives the result of the inner handler so we can detect whether
		// the timeout fired before the handler returned.
		doneCh := make(chan struct{})

		// tw captures the first WriteHeader call so we can detect a race between
		// the handler writing a response and us writing 503.
		tw := &trackingWriter{ResponseWriter: w}

		go func() {
			defer close(doneCh)
			next.ServeHTTP(tw, r.WithContext(ctx))
		}()

		select {
		case <-doneCh:
			// Handler finished in time — nothing to do.
		case <-ctx.Done():
			<-doneCh // wait for goroutine to exit before touching w
			if !tw.written {
				opts.Logger.Warn("request timed out",
					"method", r.Method,
					"path", r.URL.Path,
					"timeout", opts.Timeout.String(),
				)
				http.Error(w, "request timeout", http.StatusServiceUnavailable)
			}
		}
	})
}

// trackingWriter wraps http.ResponseWriter and records whether a status code
// has already been sent so the middleware can avoid a double-write.
type trackingWriter struct {
	http.ResponseWriter
	written bool
}

func (tw *trackingWriter) WriteHeader(code int) {
	tw.written = true
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *trackingWriter) Write(b []byte) (int, error) {
	tw.written = true
	return tw.ResponseWriter.Write(b)
}

package timeout_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andyosyndrome/retryq/internal/timeout"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func slowHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// honour cancellation — do not write
		}
	})
}

func silentOpts(d time.Duration) timeout.Options {
	return timeout.Options{
		Timeout: d,
		Logger:  slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError + 10})),
	}
}

func TestMiddleware_AllowsFastRequest(t *testing.T) {
	handler := timeout.NewMiddleware(http.HandlerFunc(okHandler), silentOpts(500*time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Returns503OnTimeout(t *testing.T) {
	slow := slowHandler(200 * time.Millisecond)
	handler := timeout.NewMiddleware(slow, silentOpts(20*time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_ZeroTimeout_UsesDefault(t *testing.T) {
	// A zero timeout should fall back to the 30 s default and not panic.
	handler := timeout.NewMiddleware(http.HandlerFunc(okHandler), timeout.Options{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_NilLogger_UsesDefault(t *testing.T) {
	// nil Logger must not panic — falls back to slog.Default().
	handler := timeout.NewMiddleware(http.HandlerFunc(okHandler), timeout.Options{
		Timeout: time.Second,
		Logger:  nil,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req) // must not panic
}

func TestDefaultOptions_SaneValues(t *testing.T) {
	opts := timeout.DefaultOptions()
	if opts.Timeout != 30*time.Second {
		t.Fatalf("expected 30s default, got %s", opts.Timeout)
	}
	if opts.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

package idempotency_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andrebq/retryq/internal/idempotency"
)

func okHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_NoKey_AlwaysAllows(t *testing.T) {
	h := idempotency.NewMiddleware(okHandler(t), nil)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("attempt %d: want 200, got %d", i, rec.Code)
		}
	}
}

func TestMiddleware_FirstRequestAllowed(t *testing.T) {
	h := idempotency.NewMiddleware(okHandler(t), nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(idempotency.DefaultHeader, "key-abc")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestMiddleware_DuplicateKeyRejected(t *testing.T) {
	h := idempotency.NewMiddleware(okHandler(t), nil)

	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(idempotency.DefaultHeader, "key-dup")
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := send(); got != http.StatusOK {
		t.Fatalf("first: want 200, got %d", got)
	}
	if got := send(); got != http.StatusConflict {
		t.Fatalf("second: want 409, got %d", got)
	}
}

func TestMiddleware_CustomHeader(t *testing.T) {
	opts := &idempotency.Options{
		Header:  "X-Request-ID",
		TTL:     time.Minute,
		MaxKeys: 100,
	}
	h := idempotency.NewMiddleware(okHandler(t), opts)

	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-Request-ID", "custom-key")
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := send(); got != http.StatusOK {
		t.Fatalf("first: want 200, got %d", got)
	}
	if got := send(); got != http.StatusConflict {
		t.Fatalf("second: want 409, got %d", got)
	}
}

func TestMiddleware_NilOpts_UsesDefaults(t *testing.T) {
	// Ensure NewMiddleware does not panic with nil opts.
	h := idempotency.NewMiddleware(okHandler(t), nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

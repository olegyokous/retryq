package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andrebq/retryq/internal/ratelimit"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_AllowsRequestUnderLimit(t *testing.T) {
	l := ratelimit.New(ratelimit.Options{Rate: 10, Burst: 10})
	mw := ratelimit.NewMiddleware(l, time.Second)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/enqueue", nil)
	mw.Handler(okHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_BlocksRequestOverLimit(t *testing.T) {
	// Burst of 1 means the second request is rejected.
	l := ratelimit.New(ratelimit.Options{Rate: 1, Burst: 1})
	mw := ratelimit.NewMiddleware(l, time.Second)

	handler := mw.Handler(okHandler())

	// First request should pass.
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, httptest.NewRequest(http.MethodPost, "/enqueue", nil))
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected first request to pass, got %d", rec1.Code)
	}

	// Second request should be rate-limited.
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodPost, "/enqueue", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}

func TestMiddleware_SetsRetryAfterHeader(t *testing.T) {
	l := ratelimit.New(ratelimit.Options{Rate: 1, Burst: 1})
	mw := ratelimit.NewMiddleware(l, 5*time.Second)
	handler := mw.Handler(okHandler())

	// Exhaust the burst.
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

	if got := rec.Header().Get(ratelimit.RetryAfterHeader); got != "5" {
		t.Fatalf("expected Retry-After: 5, got %q", got)
	}
}

func TestMiddleware_NilLimiter_AlwaysAllows(t *testing.T) {
	l := ratelimit.New(ratelimit.Options{Rate: 0, Burst: 0})
	mw := ratelimit.NewMiddleware(l, time.Second)

	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		mw.Handler(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("iteration %d: expected 200 for zero-rate limiter, got %d", i, rec.Code)
		}
	}
}

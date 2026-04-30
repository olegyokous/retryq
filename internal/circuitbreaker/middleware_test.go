package circuitbreaker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andrebq/retryq/internal/circuitbreaker"
)

func okHandlerCB(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func serverErrorHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func cbOpts() circuitbreaker.Options {
	return circuitbreaker.Options{
		MaxFailures:    2,
		CooldownPeriod: 100 * time.Millisecond,
		HalfOpenProbes: 1,
	}
}

func TestMiddleware_AllowsWhenClosed(t *testing.T) {
	cb := circuitbreaker.New(cbOpts())
	h := circuitbreaker.NewMiddleware(cb, http.HandlerFunc(okHandlerCB))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_BlocksWhenOpen(t *testing.T) {
	cb := circuitbreaker.New(cbOpts())
	h := circuitbreaker.NewMiddleware(cb, http.HandlerFunc(serverErrorHandler))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if rec.Header().Get("X-Circuit-State") != "open" {
		t.Fatalf("expected X-Circuit-State: open, got %q", rec.Header().Get("X-Circuit-State"))
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}
}

func TestMiddleware_NilCircuitBreaker_AlwaysAllows(t *testing.T) {
	h := circuitbreaker.NewMiddleware(nil, http.HandlerFunc(okHandlerCB))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_SetsCircuitStateHeader(t *testing.T) {
	cb := circuitbreaker.New(cbOpts())
	h := circuitbreaker.NewMiddleware(cb, http.HandlerFunc(okHandlerCB))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get("X-Circuit-State") != "closed" {
		t.Fatalf("expected X-Circuit-State: closed, got %q", rec.Header().Get("X-Circuit-State"))
	}
}

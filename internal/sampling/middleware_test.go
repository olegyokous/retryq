package sampling_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/retryq/internal/sampling"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestMiddleware_AllowsWhenSampled(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 1.0})
	h := sampling.NewMiddleware(s, http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_BlocksWhenNotSampled(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 0.0})
	h := sampling.NewMiddleware(s, http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

func TestMiddleware_SetsRetryAfterHeader(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 0.0})
	h := sampling.NewMiddleware(s, http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header to be set")
	}
}

func TestMiddleware_NilSampler_AlwaysAllows(t *testing.T) {
	h := sampling.NewMiddleware(nil, http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("nil sampler: expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_SetsSampleRateHeader(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 0.0})
	h := sampling.NewMiddleware(s, http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get("X-Sample-Rate") != "0.0000" {
		t.Errorf("unexpected X-Sample-Rate: %s", rec.Header().Get("X-Sample-Rate"))
	}
}

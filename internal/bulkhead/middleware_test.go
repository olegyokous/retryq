package bulkhead_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/andygeiss/retryq/internal/bulkhead"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_AllowsUnderLimit(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 5})
	h := bulkhead.NewMiddleware(b, okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_BlocksWhenFull(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 0})
	// Force full by setting max=1 and holding the slot.
	b2 := bulkhead.New(bulkhead.Options{MaxConcurrent: 1})
	_ = b2.Acquire() // hold slot

	h := bulkhead.NewMiddleware(b2, okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	b2.Release()
	_ = b // suppress unused warning
}

func TestMiddleware_SetsLimitHeader(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 42})
	h := bulkhead.NewMiddleware(b, okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("X-Bulkhead-Limit"); got != "42" {
		t.Fatalf("expected header 42, got %q", got)
	}
}

func TestMiddleware_NilBulkhead_AlwaysAllows(t *testing.T) {
	h := bulkhead.NewMiddleware(nil, okHandler())
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rec.Code)
			}
		}()
	}
	wg.Wait()
}

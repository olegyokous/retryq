package priority_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sysradium/retryq/internal/priority"
)

func okPriorityHandler(t *testing.T, want priority.Priority) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := priority.FromContext(r)
		if got != want {
			t.Errorf("FromContext = %s; want %s", got, want)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_DefaultsToNormalWhenHeaderAbsent(t *testing.T) {
	handler := priority.Middleware(okPriorityHandler(t, priority.Normal))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
}

func TestMiddleware_ParsesHighPriority(t *testing.T) {
	handler := priority.Middleware(okPriorityHandler(t, priority.High))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(priority.HeaderPriority, "high")
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
}

func TestMiddleware_ParsesLowPriority(t *testing.T) {
	handler := priority.Middleware(okPriorityHandler(t, priority.Low))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(priority.HeaderPriority, "low")
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
}

func TestMiddleware_FallsBackToNormalOnUnknownValue(t *testing.T) {
	handler := priority.Middleware(okPriorityHandler(t, priority.Normal))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(priority.HeaderPriority, "ultra-mega-high")
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
}

func TestFromContext_ReturnsNormalWhenNotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := priority.FromContext(req); got != priority.Normal {
		t.Errorf("FromContext = %s; want normal", got)
	}
}

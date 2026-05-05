package drain_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/retryq/internal/drain"
)

func TestHandler_Drained(t *testing.T) {
	w := &stubWaiter{}
	h := drain.NewHandler(w)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/drain", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp drain.StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Drained {
		t.Error("expected drained=true")
	}
	if resp.Inflight != 0 {
		t.Errorf("expected inflight=0, got %d", resp.Inflight)
	}
}

func TestHandler_NotDrained(t *testing.T) {
	w := &stubWaiter{}
	w.set(3)
	h := drain.NewHandler(w)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/drain", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp drain.StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Drained {
		t.Error("expected drained=false")
	}
	if resp.Inflight != 3 {
		t.Errorf("expected inflight=3, got %d", resp.Inflight)
	}
}

func TestHandler_WrongMethod(t *testing.T) {
	w := &stubWaiter{}
	h := drain.NewHandler(w)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/drain", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

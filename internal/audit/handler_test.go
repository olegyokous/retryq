package audit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sablierapp/retryq/internal/audit"
)

func newLog(events ...audit.Event) *audit.Log {
	l := audit.New(100)
	for _, e := range events {
		l.Record(e)
	}
	return l
}

func TestHandler_ReturnsEvents(t *testing.T) {
	l := newLog(
		audit.Event{ID: "1", Kind: audit.EventEnqueued},
		audit.Event{ID: "2", Kind: audit.EventDispatched},
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	audit.Handler(l)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if int(body["count"].(float64)) != 2 {
		t.Errorf("expected count 2, got %v", body["count"])
	}
}

func TestHandler_LimitParam(t *testing.T) {
	l := newLog(
		audit.Event{ID: "a", Kind: audit.EventRetry},
		audit.Event{ID: "b", Kind: audit.EventRetry},
		audit.Event{ID: "c", Kind: audit.EventRetry},
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit?limit=2", nil)
	audit.Handler(l)(rec, req)

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if int(body["count"].(float64)) != 2 {
		t.Errorf("expected 2 events with limit=2, got %v", body["count"])
	}
}

func TestHandler_WrongMethod(t *testing.T) {
	l := audit.New(10)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit", nil)
	audit.Handler(l)(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_EmptyLog(t *testing.T) {
	l := audit.New(10)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	audit.Handler(l)(rec, req)
	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if int(body["count"].(float64)) != 0 {
		t.Errorf("expected count 0 for empty log")
	}
}

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andyflux/retryq/internal/deadletter"
)

func TestHandleDeadLetter_EmptyStore(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/dead-letters", nil)
	rec := httptest.NewRecorder()

	srv.handleDeadLetter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []deadletter.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(entries))
	}
}

func TestHandleDeadLetter_ReturnsList(t *testing.T) {
	srv := newTestServer(t)

	entry := deadletter.Entry{
		ID:        "abc-123",
		TargetURL: "http://example.com/hook",
		Payload:   []byte(`{"event":"test"}`),
		FailedAt:  time.Now().UTC(),
		Reason:    "connection refused",
		Attempts:  3,
	}
	srv.deadLetterStore.Add(entry)

	req := httptest.NewRequest(http.MethodGet, "/dead-letters", nil)
	rec := httptest.NewRecorder()

	srv.handleDeadLetter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []deadletter.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != entry.ID {
		t.Errorf("expected ID %q, got %q", entry.ID, entries[0].ID)
	}
	if entries[0].Reason != entry.Reason {
		t.Errorf("expected reason %q, got %q", entry.Reason, entries[0].Reason)
	}
}

func TestHandleDeadLetter_WrongMethod(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/dead-letters", nil)
	rec := httptest.NewRecorder()

	srv.handleDeadLetter(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

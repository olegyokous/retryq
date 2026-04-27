package deadletter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygeiss/retryq/internal/deadletter"
)

func makeStore(t *testing.T, entries ...deadletter.Entry) *deadletter.Store {
	t.Helper()
	s := deadletter.New(10)
	for _, e := range entries {
		s.Add(e)
	}
	return s
}

func sampleEntry(id string) deadletter.Entry {
	return deadletter.Entry{
		ID:        id,
		TargetURL: "http://example.com",
		Payload:   []byte(`{"key":"value"}`),
		FailedAt:  time.Now(),
	}
}

func TestHandleRequeue_Success(t *testing.T) {
	entry := sampleEntry("abc-123")
	s := makeStore(t, entry)

	var requeued deadletter.Entry
	enqueue := func(e deadletter.Entry) error { requeued = e; return nil }

	body, _ := json.Marshal(map[string]string{"id": "abc-123"})
	req := httptest.NewRequest(http.MethodPost, "/dead-letter/requeue", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	deadletter.HandleRequeue(s, enqueue)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if requeued.ID != "abc-123" {
		t.Errorf("expected requeued ID abc-123, got %s", requeued.ID)
	}
	if s.Len() != 0 {
		t.Errorf("expected store to be empty after requeue")
	}
}

func TestHandleRequeue_NotFound(t *testing.T) {
	s := makeStore(t)
	body, _ := json.Marshal(map[string]string{"id": "missing"})
	req := httptest.NewRequest(http.MethodPost, "/dead-letter/requeue", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	deadletter.HandleRequeue(s, func(e deadletter.Entry) error { return nil })(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleRequeue_EnqueueError_RestoresEntry(t *testing.T) {
	entry := sampleEntry("xyz-999")
	s := makeStore(t, entry)

	enqueue := func(e deadletter.Entry) error { return errors.New("queue full") }

	body, _ := json.Marshal(map[string]string{"id": "xyz-999"})
	req := httptest.NewRequest(http.MethodPost, "/dead-letter/requeue", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	deadletter.HandleRequeue(s, enqueue)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if s.Len() != 1 {
		t.Errorf("expected entry to be restored in store, got len=%d", s.Len())
	}
}

func TestHandleRequeue_WrongMethod(t *testing.T) {
	s := makeStore(t)
	req := httptest.NewRequest(http.MethodGet, "/dead-letter/requeue", nil)
	rec := httptest.NewRecorder()

	deadletter.HandleRequeue(s, func(e deadletter.Entry) error { return nil })(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

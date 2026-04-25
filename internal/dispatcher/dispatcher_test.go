package dispatcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/retryq/internal/dispatcher"
	"github.com/yourorg/retryq/internal/queue"
)

func newItem(method, url string) *queue.Item {
	return &queue.Item{
		ID:        "test-id",
		Method:    method,
		TargetURL: url,
		Headers:   map[string]string{"Content-Type": "application/json"},
		Payload:   []byte(`{"hello":"world"}`),
	}
}

func TestDispatch_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	d := dispatcher.New(dispatcher.WithTimeout(2 * time.Second))
	err := d.Dispatch(context.Background(), newItem(http.MethodPost, ts.URL))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDispatch_Non2xxReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	d := dispatcher.New()
	err := d.Dispatch(context.Background(), newItem(http.MethodPost, ts.URL))
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestDispatch_TransportErrorReturnsError(t *testing.T) {
	d := dispatcher.New(dispatcher.WithTimeout(500 * time.Millisecond))
	err := d.Dispatch(context.Background(), newItem(http.MethodPost, "http://127.0.0.1:1"))
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}

func TestDispatch_ForwardsHeaders(t *testing.T) {
	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	d := dispatcher.New()
	_ = d.Dispatch(context.Background(), newItem(http.MethodPost, ts.URL))
	if gotHeader != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", gotHeader)
	}
}

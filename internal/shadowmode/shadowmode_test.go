package shadowmode_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/your-org/retryq/internal/shadowmode"
)

func TestMirror_SendsRequestToShadow(t *testing.T) {
	var mu sync.Mutex
	var received *http.Request
	var receivedBody string

	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		received = r
		receivedBody = string(body)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	opts := shadowmode.DefaultOptions()
	opts.ShadowURL = shadow.URL
	d := shadowmode.New(opts)

	req := httptest.NewRequest(http.MethodPost, "/enqueue", strings.NewReader(`{"hello":"world"}`))
	req.Header.Set("Content-Type", "application/json")

	d.Mirror(req, []byte(`{"hello":"world"}`))

	// Give the goroutine time to complete.
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if received == nil {
		t.Fatal("shadow server received no request")
	}
	if receivedBody != `{"hello":"world"}` {
		t.Errorf("unexpected body: %s", receivedBody)
	}
	if received.Header.Get("X-Shadow-Mode") != "1" {
		t.Error("X-Shadow-Mode header not set")
	}
}

func TestMirror_NoopWhenURLEmpty(t *testing.T) {
	// Should not panic or send anything.
	d := shadowmode.New(shadowmode.DefaultOptions())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	d.Mirror(req, nil) // no shadow URL configured
}

func TestMiddleware_PassesThroughToPrimary(t *testing.T) {
	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write(body)
	})

	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	opts := shadowmode.DefaultOptions()
	opts.ShadowURL = shadow.URL
	d := shadowmode.New(opts)

	handler := shadowmode.NewMiddleware(d, primary)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/enqueue", strings.NewReader(`payload`))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	if rec.Body.String() != "payload" {
		t.Errorf("primary did not receive body, got: %s", rec.Body.String())
	}
}

func TestMiddleware_NilDispatcher_PassesThrough(t *testing.T) {
	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := shadowmode.NewMiddleware(nil, primary)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

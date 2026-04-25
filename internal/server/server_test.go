package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/retryq/retryq/internal/config"
	"github.com/retryq/retryq/internal/metrics"
	"github.com/retryq/retryq/internal/queue"
	"github.com/retryq/retryq/internal/server"
	"github.com/prometheus/client_golang/prometheus"
)

func newTestServer(t *testing.T) *server.Server {
	t.Helper()
	cfg := config.Default()
	q := queue.New(cfg)
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)
	return server.New(":0", q, m)
}

func TestHandleEnqueue_Success(t *testing.T) {
	srv := newTestServer(t)
	body := `{"target_url":"http://example.com","method":"POST"}`
	req := httptest.NewRequest(http.MethodPost, "/enqueue", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["id"] == "" {
		t.Fatal("expected non-empty id in response")
	}
}

func TestHandleEnqueue_MissingTargetURL(t *testing.T) {
	srv := newTestServer(t)
	body := `{"method":"POST"}`
	req := httptest.NewRequest(http.MethodPost, "/enqueue", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleEnqueue_WrongMethod(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/enqueue", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleEnqueue_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

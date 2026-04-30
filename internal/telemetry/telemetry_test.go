package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrebq/retryq/internal/telemetry"
)

func TestNew_GeneratesUniqueIDs(t *testing.T) {
	a, err := telemetry.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := telemetry.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == b {
		t.Errorf("expected unique IDs, got duplicate: %q", a)
	}
	if len(a) != 32 {
		t.Errorf("expected 32-char hex string, got len %d", len(a))
	}
}

func TestWithContext_RoundTrip(t *testing.T) {
	id := telemetry.MustNew()
	ctx := telemetry.WithContext(context.Background(), id)
	got, ok := telemetry.FromContext(ctx)
	if !ok {
		t.Fatal("expected trace ID in context")
	}
	if got != id {
		t.Errorf("got %q, want %q", got, id)
	}
}

func TestFromContext_Missing(t *testing.T) {
	_, ok := telemetry.FromContext(context.Background())
	if ok {
		t.Error("expected no trace ID in empty context")
	}
}

func TestMiddleware_GeneratesIDWhenAbsent(t *testing.T) {
	var capturedID telemetry.TraceID
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedID, _ = telemetry.FromContext(r.Context())
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	telemetry.Middleware(inner).ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected trace ID in context")
	}
	if rec.Header().Get(telemetry.Header) != string(capturedID) {
		t.Errorf("response header mismatch: got %q, want %q",
			rec.Header().Get(telemetry.Header), capturedID)
	}
}

func TestMiddleware_PropagatesIncomingID(t *testing.T) {
	incoming := telemetry.MustNew()
	var capturedID telemetry.TraceID
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedID, _ = telemetry.FromContext(r.Context())
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(telemetry.Header, string(incoming))
	telemetry.Middleware(inner).ServeHTTP(rec, req)

	if capturedID != incoming {
		t.Errorf("got %q, want %q", capturedID, incoming)
	}
}

func TestLogAttr_Key(t *testing.T) {
	id := telemetry.MustNew()
	attr := telemetry.LogAttr(id)
	if attr.Key != "trace_id" {
		t.Errorf("unexpected key %q", attr.Key)
	}
	if attr.Value.String() != string(id) {
		t.Errorf("unexpected value %q", attr.Value.String())
	}
}

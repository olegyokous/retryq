package worker_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/example/retryq/internal/config"
	"github.com/example/retryq/internal/metrics"
	"github.com/example/retryq/internal/queue"
	"github.com/example/retryq/internal/worker"
	"github.com/prometheus/client_golang/prometheus"
)

func newDeps(t *testing.T) (*queue.Queue, *metrics.Metrics) {
	t.Helper()
	cfg := config.Default()
	q := queue.New(cfg)
	m := metrics.New(prometheus.NewRegistry())
	return q, m
}

func TestWorker_SuccessfulDispatch(t *testing.T) {
	q, m := newDeps(t)

	item := &queue.Item{ID: "abc", Payload: []byte(`{}`), Status: queue.StatusPending}
	q.Enqueue(item)

	dispatcher := func(_ context.Context, _ *queue.Item) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK}, nil
	}

	w := worker.New(q, m, dispatcher, 10*time.Millisecond, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	if got := q.Len(); got != 0 {
		t.Errorf("expected queue empty after success, got %d items", got)
	}
}

func TestWorker_FailedDispatchSchedulesRetry(t *testing.T) {
	q, m := newDeps(t)

	item := &queue.Item{ID: "xyz", Payload: []byte(`{}`), Status: queue.StatusPending}
	q.Enqueue(item)

	dispatcher := func(_ context.Context, _ *queue.Item) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusInternalServerError}, nil
	}

	w := worker.New(q, m, dispatcher, 10*time.Millisecond, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	// Item should still be in the queue (retrying or dead-letter)
	if got := q.Len(); got == 0 {
		t.Error("expected item to remain in queue after failure")
	}
}

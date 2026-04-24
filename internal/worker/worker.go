// Package worker provides a background processor that drains the retry queue,
// dispatching ready items to a configurable HTTP handler and recording outcomes.
package worker

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/retryq/internal/metrics"
	"github.com/example/retryq/internal/queue"
)

// Dispatcher is the function signature used to forward a queued item.
type Dispatcher func(ctx context.Context, item *queue.Item) (*http.Response, error)

// Worker polls the queue and dispatches ready items.
type Worker struct {
	q        *queue.Queue
	m        *metrics.Metrics
	dispatch Dispatcher
	pollInterval time.Duration
	log      *slog.Logger
}

// New creates a Worker with the given queue, metrics, dispatcher, and poll interval.
func New(q *queue.Queue, m *metrics.Metrics, d Dispatcher, pollInterval time.Duration, log *slog.Logger) *Worker {
	if log == nil {
		log = slog.Default()
	}
	return &Worker{
		q:            q,
		m:            m,
		dispatch:     d,
		pollInterval: pollInterval,
		log:          log,
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("worker shutting down")
			return
		case <-ticker.C:
			w.processReady(ctx)
		}
	}
}

func (w *Worker) processReady(ctx context.Context) {
	items := w.q.DrainReady()
	for _, item := range items {
		w.log.Debug("dispatching item", "id", item.ID, "attempt", item.Attempts)
		resp, err := w.dispatch(ctx, item)
		if err != nil || (resp != nil && resp.StatusCode >= 500) {
			status := 0
			if resp != nil {
				status = resp.StatusCode
			}
			w.log.Warn("dispatch failed", "id", item.ID, "status", status, "err", err)
			w.q.RecordFailure(item)
			w.m.RetryScheduled.Inc()
		} else {
			w.log.Info("dispatch succeeded", "id", item.ID)
			w.m.Succeeded.Inc()
			w.m.QueueDepth.Dec()
		}
	}
}

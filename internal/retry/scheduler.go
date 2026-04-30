// Package retry provides scheduling logic for re-enqueuing items
// after a transient failure, wiring together the backoff policy,
// queue, and metrics layers.
package retry

import (
	"context"
	"log/slog"
	"time"

	"github.com/example/retryq/internal/backoff"
	"github.com/example/retryq/internal/metrics"
	"github.com/example/retryq/internal/queue"
)

// Enqueuer is the subset of queue.Queue used by the Scheduler.
type Enqueuer interface {
	RecordFailure(item *queue.Item) error
}

// Scheduler decides when and whether to re-schedule a failed item.
type Scheduler struct {
	q      Enqueuer
	policy backoff.Policy
	m      *metrics.Metrics
	log    *slog.Logger
}

// New returns a Scheduler wired to the provided dependencies.
func New(q Enqueuer, p backoff.Policy, m *metrics.Metrics, log *slog.Logger) *Scheduler {
	if log == nil {
		log = slog.Default()
	}
	return &Scheduler{q: q, policy: p, m: m, log: log}
}

// Schedule records a failure for item and re-enqueues it unless the
// item has been exhausted (moved to dead-letter by RecordFailure).
func (s *Scheduler) Schedule(ctx context.Context, item *queue.Item) error {
	delay := s.policy.Next(item.Attempts)
	s.log.InfoContext(ctx, "scheduling retry",
		"id", item.ID,
		"attempts", item.Attempts,
		"delay", delay.Round(time.Millisecond),
	)

	if err := s.q.RecordFailure(item); err != nil {
		s.log.ErrorContext(ctx, "record failure", "id", item.ID, "err", err)
		return err
	}

	if s.m != nil {
		s.m.Retries.Inc()
	}
	return nil
}

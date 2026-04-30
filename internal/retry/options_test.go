package retry_test

import (
	"testing"

	"github.com/example/retryq/internal/config"
	"github.com/example/retryq/internal/retry"
)

func TestDefaultOptions_HasLogger(t *testing.T) {
	opts := retry.DefaultOptions()
	if opts.Logger == nil {
		t.Fatal("expected non-nil default logger")
	}
}

func TestNewFromConfig_ReturnsScheduler(t *testing.T) {
	cfg := config.Default()
	enq := &stubEnqueuer{}
	opts := retry.DefaultOptions()

	sched := retry.NewFromConfig(enq, cfg, opts)
	if sched == nil {
		t.Fatal("expected non-nil Scheduler")
	}
}

func TestNewFromConfig_NilLogger_FallsBackToDefault(t *testing.T) {
	cfg := config.Default()
	enq := &stubEnqueuer{}
	opts := retry.Options{Logger: nil, Metrics: nil}

	// Must not panic.
	sched := retry.NewFromConfig(enq, cfg, opts)
	if sched == nil {
		t.Fatal("expected non-nil Scheduler")
	}
}

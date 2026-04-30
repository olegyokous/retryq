package retry

import (
	"log/slog"

	"github.com/example/retryq/internal/backoff"
	"github.com/example/retryq/internal/config"
	"github.com/example/retryq/internal/metrics"
)

// Options holds all optional overrides for building a Scheduler via
// NewFromConfig.
type Options struct {
	Metrics *metrics.Metrics
	Logger  *slog.Logger
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Logger: slog.Default(),
	}
}

// NewFromConfig is a convenience constructor that builds the backoff
// policy from cfg and delegates to New.
func NewFromConfig(q Enqueuer, cfg config.Config, opts Options) *Scheduler {
	pol := backoff.New(cfg)
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return New(q, pol, opts.Metrics, opts.Logger)
}

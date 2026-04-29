// Package deadletter provides storage and management of failed queue items
// that have exhausted all retry attempts.
package deadletter

import (
	"context"
	"log/slog"
	"time"
)

// ExpiryWorker periodically purges expired dead-letter entries from the store.
type ExpiryWorker struct {
	store   *Store
	opts    ExpiryOptions
	ticker  *time.Ticker
	logger  *slog.Logger
}

// NewExpiryWorker creates an ExpiryWorker that will run Purge on the given
// store at the interval specified in opts.Interval.
// If logger is nil, slog.Default() is used.
func NewExpiryWorker(store *Store, opts ExpiryOptions, logger *slog.Logger) *ExpiryWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &ExpiryWorker{
		store:  store,
		opts:   opts,
		logger: logger,
	}
}

// Run starts the expiry loop. It blocks until ctx is cancelled.
// On each tick it calls Purge and logs the number of evicted entries.
func (w *ExpiryWorker) Run(ctx context.Context) {
	if w.opts.Interval <= 0 {
		w.logger.Info("expiry worker disabled: interval is zero or negative")
		<-ctx.Done()
		return
	}

	w.ticker = time.NewTicker(w.opts.Interval)
	defer w.ticker.Stop()

	w.logger.Info("expiry worker started",
		"interval", w.opts.Interval,
		"max_age", w.opts.MaxAge,
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("expiry worker stopped")
			return
		case <-w.ticker.C:
			w.purge()
		}
	}
}

// purge runs a single Purge cycle and logs the outcome.
func (w *ExpiryWorker) purge() {
	removed := Purge(w.store, w.opts)
	if removed > 0 {
		w.logger.Info("dead-letter expiry: purged entries",
			"count", removed,
			"max_age", w.opts.MaxAge,
		)
	}
}

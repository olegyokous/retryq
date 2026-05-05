// Package drain provides a graceful shutdown drain that waits for
// in-flight queue items to complete before the process exits.
package drain

import (
	"context"
	"log"
	"time"
)

// Waiter is satisfied by anything that can report how many items are
// currently being processed (e.g. the bulkhead or the worker pool).
type Waiter interface {
	Inflight() int
}

// Options controls drain behaviour.
type Options struct {
	// Timeout is the maximum time to wait for in-flight items to finish.
	// Defaults to 30 s.
	Timeout time.Duration
	// PollInterval is how often the drain checks the inflight count.
	// Defaults to 100 ms.
	PollInterval time.Duration
	// Logger receives progress messages. If nil, log.Default() is used.
	Logger *log.Logger
}

// DefaultOptions returns safe defaults.
func DefaultOptions() Options {
	return Options{
		Timeout:      30 * time.Second,
		PollInterval: 100 * time.Millisecond,
		Logger:       log.Default(),
	}
}

// Wait blocks until all in-flight items reported by w have finished or
// the deadline in opts.Timeout is reached.  It returns true when the
// queue drained cleanly and false when the timeout was exceeded.
func Wait(ctx context.Context, w Waiter, opts Options) bool {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultOptions().Timeout
	}
	if opts.PollInterval <= 0 {
		opts.PollInterval = DefaultOptions().PollInterval
	}
	logger := opts.Logger
	if logger == nil {
		logger = log.Default()
	}

	deadline := time.Now().Add(opts.Timeout)
	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	for {
		if w.Inflight() == 0 {
			logger.Println("drain: all in-flight items finished")
			return true
		}
		if time.Now().After(deadline) {
			logger.Printf("drain: timeout exceeded with %d item(s) still in-flight", w.Inflight())
			return false
		}
		select {
		case <-ctx.Done():
			logger.Printf("drain: context cancelled with %d item(s) still in-flight", w.Inflight())
			return false
		case <-ticker.C:
			logger.Printf("drain: waiting — %d item(s) in-flight", w.Inflight())
		}
	}
}

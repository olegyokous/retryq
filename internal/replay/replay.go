// Package replay provides a mechanism to re-dispatch dead-letter entries
// through the original HTTP pipeline with configurable concurrency.
package replay

import (
	"context"
	"log"
	"sync"

	"github.com/your-org/retryq/internal/deadletter"
)

// Dispatcher is the interface used to re-send a dead-letter entry.
type Dispatcher interface {
	Dispatch(ctx context.Context, entry deadletter.Entry) error
}

// Options controls replay behaviour.
type Options struct {
	// Concurrency is the number of entries replayed in parallel.
	// Defaults to 1 (sequential).
	Concurrency int
	Logger      *log.Logger
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Concurrency: 1,
		Logger:      log.Default(),
	}
}

// Result captures the outcome of replaying a single entry.
type Result struct {
	Entry deadletter.Entry
	Err   error
}

// Replay dispatches every entry in the slice concurrently (up to opts.Concurrency)
// and returns one Result per entry in the order they complete.
func Replay(ctx context.Context, entries []deadletter.Entry, d Dispatcher, opts Options) []Result {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}
	if opts.Logger == nil {
		opts.Logger = log.Default()
	}

	sem := make(chan struct{}, opts.Concurrency)
	results := make([]Result, len(entries))
	var wg sync.WaitGroup

	for i, e := range entries {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, entry deadletter.Entry) {
			defer wg.Done()
			defer func() { <-sem }()

			err := d.Dispatch(ctx, entry)
			if err != nil {
				opts.Logger.Printf("replay: dispatch failed for entry %s: %v", entry.ID, err)
			}
			results[idx] = Result{Entry: entry, Err: err}
		}(i, e)
	}

	wg.Wait()
	return results
}

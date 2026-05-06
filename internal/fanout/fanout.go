// Package fanout provides a mechanism to dispatch a single retry-queue item
// to multiple target URLs concurrently, collecting per-target results.
package fanout

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Dispatcher is the interface used to forward a request to a single target.
type Dispatcher interface {
	Dispatch(ctx context.Context, targetURL string, headers http.Header, body []byte) error
}

// Result holds the outcome of dispatching to one target URL.
type Result struct {
	TargetURL string
	Err       error
	Duration  time.Duration
}

// Options controls fanout behaviour.
type Options struct {
	// Timeout is applied per-target dispatch. Zero means no per-target timeout.
	Timeout time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Timeout: 5 * time.Second,
	}
}

// Fan dispatches body to every URL in targets concurrently and returns one
// Result per target. The returned slice preserves the order of targets.
func Fan(
	ctx context.Context,
	d Dispatcher,
	targets []string,
	headers http.Header,
	body []byte,
	opts Options,
) []Result {
	if len(targets) == 0 {
		return nil
	}

	results := make([]Result, len(targets))
	var wg sync.WaitGroup

	for i, url := range targets {
		wg.Add(1)
		go func(idx int, target string) {
			defer wg.Done()

			dispatchCtx := ctx
			if opts.Timeout > 0 {
				var cancel context.CancelFunc
				dispatchCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
				defer cancel()
			}

			start := time.Now()
			err := d.Dispatch(dispatchCtx, target, headers, body)
			results[idx] = Result{
				TargetURL: target,
				Err:       err,
				Duration:  time.Since(start),
			}
		}(i, url)
	}

	wg.Wait()
	return results
}

// FirstError returns the first non-nil error found in results, or nil when all
// targets succeeded. Useful for callers that treat any failure as a retry.
func FirstError(results []Result) error {
	for _, r := range results {
		if r.Err != nil {
			return fmt.Errorf("fanout target %q: %w", r.TargetURL, r.Err)
		}
	}
	return nil
}

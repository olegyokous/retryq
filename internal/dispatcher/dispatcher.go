// Package dispatcher provides an HTTP dispatcher that forwards queue items
// to their target endpoints and reports outcomes back to the queue.
package dispatcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourorg/retryq/internal/queue"
)

// Dispatcher sends queued items to their target HTTP endpoints.
type Dispatcher struct {
	client  *http.Client
	timeout time.Duration
}

// Option is a functional option for Dispatcher.
type Option func(*Dispatcher)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(dp *Dispatcher) {
		dp.timeout = d
		dp.client.Timeout = d
	}
}

// WithClient replaces the underlying HTTP client.
func WithClient(c *http.Client) Option {
	return func(dp *Dispatcher) {
		dp.client = c
	}
}

// New creates a Dispatcher with sensible defaults.
func New(opts ...Option) *Dispatcher {
	d := &Dispatcher{
		client:  &http.Client{Timeout: 10 * time.Second},
		timeout: 10 * time.Second,
	}
	for _, o := range opts {
		o(d)
	}
	return d
}

// Dispatch sends the item's payload to its target URL using its method.
// It returns a non-nil error for any non-2xx response or transport failure.
func (d *Dispatcher) Dispatch(ctx context.Context, item *queue.Item) error {
	req, err := http.NewRequestWithContext(ctx, item.Method, item.TargetURL, item.Body())
	if err != nil {
		return fmt.Errorf("dispatcher: build request: %w", err)
	}
	for k, v := range item.Headers {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("dispatcher: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dispatcher: unexpected status %d", resp.StatusCode)
	}
	return nil
}

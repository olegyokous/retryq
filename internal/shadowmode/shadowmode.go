// Package shadowmode provides a shadow-mode dispatcher that forwards
// requests to a secondary target without affecting the primary response.
// This is useful for traffic mirroring, dark launches, and A/B testing.
package shadowmode

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Options configures the shadow dispatcher.
type Options struct {
	// ShadowURL is the base URL to mirror requests to.
	ShadowURL string
	// Timeout is the maximum time to wait for the shadow request.
	Timeout time.Duration
	// Logger is used to record shadow dispatch outcomes.
	Logger *slog.Logger
	// Client is the HTTP client used for shadow requests.
	Client *http.Client
}

// DefaultOptions returns a sensible default Options.
func DefaultOptions() Options {
	return Options{
		Timeout: 2 * time.Second,
		Logger:  slog.Default(),
		Client:  &http.Client{Timeout: 2 * time.Second},
	}
}

// Dispatcher mirrors incoming HTTP requests to a shadow target.
type Dispatcher struct {
	opts Options
}

// New creates a new shadow Dispatcher with the given options.
func New(opts Options) *Dispatcher {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: opts.Timeout}
	}
	if opts.Timeout == 0 {
		opts.Timeout = 2 * time.Second
	}
	return &Dispatcher{opts: opts}
}

// Mirror asynchronously forwards r to the configured shadow URL.
// The caller's request body is consumed safely via bodyBytes.
func (d *Dispatcher) Mirror(r *http.Request, bodyBytes []byte) {
	if d.opts.ShadowURL == "" {
		return
	}
	go d.send(r, bodyBytes)
}

func (d *Dispatcher) send(orig *http.Request, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), d.opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, orig.Method, d.opts.ShadowURL, bytes.NewReader(body))
	if err != nil {
		d.opts.Logger.Warn("shadow: failed to build request", "error", err)
		return
	}

	for key, vals := range orig.Header {
		for _, v := range vals {
			req.Header.Add(key, v)
		}
	}
	req.Header.Set("X-Shadow-Mode", "1")

	resp, err := d.opts.Client.Do(req)
	if err != nil {
		d.opts.Logger.Warn("shadow: request failed", "error", err, "url", d.opts.ShadowURL)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	d.opts.Logger.Debug("shadow: request mirrored", "status", resp.StatusCode, "url", d.opts.ShadowURL)
}

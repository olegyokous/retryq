// Package bulkhead implements a concurrency-limiting bulkhead pattern
// that caps the number of in-flight requests to protect downstream services.
package bulkhead

import (
	"errors"
	"sync/atomic"
)

// ErrFull is returned when the bulkhead has reached its concurrency limit.
var ErrFull = errors.New("bulkhead: at capacity")

// Options configures a Bulkhead.
type Options struct {
	// MaxConcurrent is the maximum number of simultaneous in-flight requests.
	MaxConcurrent int64
}

// DefaultOptions returns sane defaults for a Bulkhead.
func DefaultOptions() Options {
	return Options{
		MaxConcurrent: 100,
	}
}

// Bulkhead limits the number of concurrent operations.
type Bulkhead struct {
	max     int64
	inflight atomic.Int64
}

// New creates a new Bulkhead with the given options.
// If opts.MaxConcurrent <= 0 the bulkhead is disabled (always allows).
func New(opts Options) *Bulkhead {
	return &Bulkhead{max: opts.MaxConcurrent}
}

// Acquire attempts to acquire a slot. It returns ErrFull when at capacity.
// Callers must call Release exactly once after a successful Acquire.
func (b *Bulkhead) Acquire() error {
	if b.max <= 0 {
		return nil
	}
	current := b.inflight.Add(1)
	if current > b.max {
		b.inflight.Add(-1)
		return ErrFull
	}
	return nil
}

// Release frees a previously acquired slot.
func (b *Bulkhead) Release() {
	if b.max <= 0 {
		return
	}
	b.inflight.Add(-1)
}

// Inflight returns the current number of in-flight operations.
func (b *Bulkhead) Inflight() int64 {
	return b.inflight.Load()
}

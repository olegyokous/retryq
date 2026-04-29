// Package ratelimit provides a token-bucket rate limiter for controlling
// how frequently retry queue items are dispatched to target endpoints.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a simple token-bucket rate limiter.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	now      func() time.Time
}

// Options configures the rate limiter.
type Options struct {
	// Rate is the number of requests allowed per second.
	Rate float64
	// Burst is the maximum number of tokens that can accumulate.
	Burst float64
}

// DefaultOptions returns sensible defaults for the rate limiter.
func DefaultOptions() Options {
	return Options{
		Rate:  100,
		Burst: 100,
	}
}

// New creates a new Limiter with the given options.
// Returns nil if rate <= 0 (unlimited).
func New(opts Options) *Limiter {
	if opts.Rate <= 0 {
		return nil
	}
	if opts.Burst <= 0 {
		opts.Burst = opts.Rate
	}
	return &Limiter{
		tokens:   opts.Burst,
		max:      opts.Burst,
		rate:     opts.Rate,
		lastTick: time.Now(),
		now:      time.Now,
	}
}

// Allow reports whether one token may be consumed.
// It refills tokens based on elapsed time since the last call.
func (l *Limiter) Allow() bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.lastTick = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

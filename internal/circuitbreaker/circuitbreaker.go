// Package circuitbreaker provides a simple per-target circuit breaker that
// opens after a configurable number of consecutive failures and resets after
// a cooldown period, preventing the worker from hammering unhealthy endpoints.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit is open and requests are rejected.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the current state of a circuit breaker.
type State int

const (
	StateClosed State = iota
	StateOpen
)

// Breaker is a per-target circuit breaker.
type Breaker struct {
	mu           sync.Mutex
	failures     int
	maxFailures  int
	cooldown     time.Duration
	openedAt     time.Time
	state        State
}

// Options configures a Breaker.
type Options struct {
	// MaxFailures is the number of consecutive failures before the circuit opens.
	MaxFailures int
	// Cooldown is how long the circuit stays open before allowing a probe.
	Cooldown time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxFailures: 5,
		Cooldown:    30 * time.Second,
	}
}

// New creates a new Breaker with the given options.
func New(opts Options) *Breaker {
	if opts.MaxFailures <= 0 {
		opts.MaxFailures = DefaultOptions().MaxFailures
	}
	if opts.Cooldown <= 0 {
		opts.Cooldown = DefaultOptions().Cooldown
	}
	return &Breaker{
		maxFailures: opts.MaxFailures,
		cooldown:    opts.Cooldown,
		state:       StateClosed,
	}
}

// Allow returns nil if the request may proceed, or ErrOpen if the circuit is
// open. A half-open probe is permitted once the cooldown elapses.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateOpen {
		if time.Since(b.openedAt) >= b.cooldown {
			// Allow a single probe attempt; reset failure count.
			b.state = StateClosed
			b.failures = 0
		} else {
			return ErrOpen
		}
	}
	return nil
}

// RecordSuccess resets the failure counter and closes the circuit.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit when the
// threshold is reached.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.maxFailures {
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}

// State returns the current state of the breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

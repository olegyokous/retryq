// Package sampling provides probabilistic and rate-based sampling for
// controlling which retry events are recorded in the audit log and
// emitted as trace spans.
package sampling

import (
	"math/rand"
	"sync"
)

// Sampler decides whether a given event should be sampled.
type Sampler struct {
	mu   sync.Mutex
	rate float64 // 0.0 – 1.0
	rng  *rand.Rand
}

// Options configures a Sampler.
type Options struct {
	// Rate is the fraction of events to sample (0.0 = none, 1.0 = all).
	Rate float64
	// Seed is used to initialise the RNG; 0 uses a fixed default.
	Seed int64
}

// DefaultOptions returns a Sampler that records every event.
func DefaultOptions() Options {
	return Options{Rate: 1.0, Seed: 42}
}

// New constructs a Sampler from opts.
// A Rate outside [0,1] is clamped to the nearest bound.
func New(opts Options) *Sampler {
	if opts.Rate < 0 {
		opts.Rate = 0
	}
	if opts.Rate > 1 {
		opts.Rate = 1
	}
	seed := opts.Seed
	if seed == 0 {
		seed = 42
	}
	return &Sampler{
		rate: opts.Rate,
		rng:  rand.New(rand.NewSource(seed)), //nolint:gosec
	}
}

// Sample returns true if the event should be recorded.
func (s *Sampler) Sample() bool {
	if s == nil {
		return true
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.rate
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() float64 {
	if s == nil {
		return 1.0
	}
	return s.rate
}

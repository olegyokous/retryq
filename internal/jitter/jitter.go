// Package jitter provides utilities for adding randomised jitter to
// retry backoff durations, reducing thundering-herd effects when many
// clients retry simultaneously.
package jitter

import (
	"math/rand"
	"time"
)

// Strategy defines how jitter is applied to a base duration.
type Strategy int

const (
	// Full applies a uniformly random duration in [0, base].
	Full Strategy = iota
	// Equal applies a duration in [base/2, base].
	Equal
	// None returns the base duration unchanged.
	None
)

// Options configures the jitter source.
type Options struct {
	// Strategy controls the jitter algorithm. Defaults to Full.
	Strategy Strategy
	// Rand is the random source. If nil, a default source is used.
	Rand *rand.Rand
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Strategy: Full,
	}
}

// Applier adds jitter to durations.
type Applier struct {
	opts Options
	rng  *rand.Rand
}

// New creates an Applier with the given options.
func New(opts Options) *Applier {
	rng := opts.Rand
	if rng == nil {
		//nolint:gosec // non-cryptographic use
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &Applier{opts: opts, rng: rng}
}

// Apply returns a jittered version of base according to the configured strategy.
func (a *Applier) Apply(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	switch a.opts.Strategy {
	case Full:
		return time.Duration(a.rng.Int63n(int64(base) + 1))
	case Equal:
		half := base / 2
		return half + time.Duration(a.rng.Int63n(int64(base-half)+1))
	default:
		return base
	}
}

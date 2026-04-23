// Package backoff computes exponential backoff durations from a RetryPolicy.
package backoff

import (
	"math"
	"math/rand"
	"time"

	"github.com/yourorg/retryq/internal/config"
)

// Calculator computes wait durations for successive retry attempts.
type Calculator struct {
	policy config.RetryPolicy
	rng    *rand.Rand
}

// New creates a Calculator seeded with the current time.
func New(policy config.RetryPolicy) *Calculator {
	return &Calculator{
		policy: policy,
		//nolint:gosec // weak RNG is acceptable for jitter
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Next returns the backoff duration for the given attempt number (1-based).
// It returns 0 and false when attempts are exhausted.
func (c *Calculator) Next(attempt int) (time.Duration, bool) {
	if attempt > c.policy.MaxAttempts {
		return 0, false
	}

	// base = InitialInterval * Multiplier^(attempt-1)
	base := float64(c.policy.InitialInterval) * math.Pow(c.policy.Multiplier, float64(attempt-1))

	// Apply jitter: base * (1 ± jitter)
	if c.policy.Jitter > 0 {
		delta := base * c.policy.Jitter
		base = base - delta + c.rng.Float64()*2*delta
	}

	d := time.Duration(base)
	if d > c.policy.MaxInterval {
		d = c.policy.MaxInterval
	}
	if d < 0 {
		d = 0
	}
	return d, true
}

// IsExhausted reports whether the given attempt number exceeds MaxAttempts.
func (c *Calculator) IsExhausted(attempt int) bool {
	return attempt > c.policy.MaxAttempts
}

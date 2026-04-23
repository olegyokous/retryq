package config

import (
	"errors"
	"time"
)

// RetryPolicy defines the exponential backoff parameters for the retry queue.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts before moving to the dead-letter queue.
	MaxAttempts int `json:"max_attempts"`
	// InitialInterval is the wait time before the first retry.
	InitialInterval time.Duration `json:"initial_interval"`
	// Multiplier is applied to the interval after each failed attempt.
	Multiplier float64 `json:"multiplier"`
	// MaxInterval caps the computed backoff duration.
	MaxInterval time.Duration `json:"max_interval"`
	// Jitter adds randomness (0.0–1.0) to avoid thundering herd.
	Jitter float64 `json:"jitter"`
}

// Config holds the top-level retryq configuration.
type Config struct {
	Retry           RetryPolicy `json:"retry"`
	DeadLetterQueue string      `json:"dead_letter_queue"`
	Workers         int         `json:"workers"`
	QueueCapacity   int         `json:"queue_capacity"`
}

// Default returns a Config populated with sensible defaults.
func Default() Config {
	return Config{
		Retry: RetryPolicy{
			MaxAttempts:     5,
			InitialInterval: 500 * time.Millisecond,
			Multiplier:      2.0,
			MaxInterval:     30 * time.Second,
			Jitter:          0.2,
		},
		DeadLetterQueue: "dead_letter",
		Workers:         4,
		QueueCapacity:   1024,
	}
}

// Validate checks that the Config values are within acceptable ranges.
func (c *Config) Validate() error {
	if c.Retry.MaxAttempts < 1 {
		return errors.New("config: max_attempts must be at least 1")
	}
	if c.Retry.InitialInterval <= 0 {
		return errors.New("config: initial_interval must be positive")
	}
	if c.Retry.Multiplier < 1.0 {
		return errors.New("config: multiplier must be >= 1.0")
	}
	if c.Retry.MaxInterval < c.Retry.InitialInterval {
		return errors.New("config: max_interval must be >= initial_interval")
	}
	if c.Retry.Jitter < 0 || c.Retry.Jitter > 1 {
		return errors.New("config: jitter must be between 0.0 and 1.0")
	}
	if c.Workers < 1 {
		return errors.New("config: workers must be at least 1")
	}
	if c.QueueCapacity < 1 {
		return errors.New("config: queue_capacity must be at least 1")
	}
	return nil
}

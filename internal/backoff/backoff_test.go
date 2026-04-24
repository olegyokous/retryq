package backoff_test

import (
	"testing"
	"time"

	"github.com/yourorg/retryq/internal/backoff"
	"github.com/yourorg/retryq/internal/config"
)

func policy() config.RetryPolicy {
	return config.RetryPolicy{
		MaxAttempts:     5,
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     10 * time.Second,
		Jitter:          0, // deterministic for tests
	}
}

func TestNext_FirstAttempt(t *testing.T) {
	calc := backoff.New(policy())
	d, ok := calc.Next(1)
	if !ok {
		t.Fatal("expected ok=true for attempt 1")
	}
	if d != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", d)
	}
}

func TestNext_Doubles(t *testing.T) {
	calc := backoff.New(policy())
	prev, _ := calc.Next(1)
	for attempt := 2; attempt <= 4; attempt++ {
		d, ok := calc.Next(attempt)
		if !ok {
			t.Fatalf("expected ok=true for attempt %d", attempt)
		}
		if d != prev*2 {
			t.Errorf("attempt %d: expected %v, got %v", attempt, prev*2, d)
		}
		prev = d
	}
}

func TestNext_CappedAtMaxInterval(t *testing.T) {
	p := policy()
	p.MaxInterval = 200 * time.Millisecond
	calc := backoff.New(p)
	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		d, _ := calc.Next(attempt)
		if d > p.MaxInterval {
			t.Errorf("attempt %d: duration %v exceeds max_interval %v", attempt, d, p.MaxInterval)
		}
	}
}

func TestNext_Exhausted(t *testing.T) {
	calc := backoff.New(policy())
	_, ok := calc.Next(6)
	if ok {
		t.Fatal("expected ok=false when attempts exhausted")
	}
}

func TestNext_AttemptZero(t *testing.T) {
	// Attempt 0 is invalid and should be treated as exhausted.
	calc := backoff.New(policy())
	_, ok := calc.Next(0)
	if ok {
		t.Fatal("expected ok=false for attempt 0")
	}
}

func TestIsExhausted(t *testing.T) {
	calc := backoff.New(policy())
	if calc.IsExhausted(5) {
		t.Error("attempt 5 should not be exhausted (MaxAttempts=5)")
	}
	if !calc.IsExhausted(6) {
		t.Error("attempt 6 should be exhausted")
	}
}

func TestNext_WithJitter_NonNegative(t *testing.T) {
	p := policy()
	p.Jitter = 0.5
	calc := backoff.New(p)
	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		d, ok := calc.Next(attempt)
		if !ok {
			t.Fatalf("unexpected exhaustion at attempt %d", attempt)
		}
		if d < 0 {
			t.Errorf("attempt %d: negative duration %v", attempt, d)
		}
	}
}

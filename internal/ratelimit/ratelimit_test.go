package ratelimit_test

import (
	"testing"
	"time"

	"github.com/iamNilotpal/retryq/internal/ratelimit"
)

func TestAllow_NilLimiter_AlwaysAllows(t *testing.T) {
	var l *ratelimit.Limiter
	for i := 0; i < 10; i++ {
		if !l.Allow() {
			t.Fatal("nil limiter should always allow")
		}
	}
}

func TestAllow_ZeroRate_ReturnsNil(t *testing.T) {
	l := ratelimit.New(ratelimit.Options{Rate: 0, Burst: 10})
	if l != nil {
		t.Fatal("expected nil limiter for zero rate")
	}
}

func TestAllow_NegativeRate_ReturnsNil(t *testing.T) {
	l := ratelimit.New(ratelimit.Options{Rate: -5, Burst: 10})
	if l != nil {
		t.Fatal("expected nil limiter for negative rate")
	}
}

func TestAllow_BurstConsumption(t *testing.T) {
	l := ratelimit.New(ratelimit.Options{Rate: 1, Burst: 3})
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}

	// Should allow burst of 3
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("expected allow on call %d", i+1)
		}
	}
	// 4th call should be denied
	if l.Allow() {
		t.Fatal("expected deny after burst exhausted")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	now := time.Now()
	opts := ratelimit.Options{Rate: 10, Burst: 1}
	l := ratelimit.New(opts)
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}

	// Consume the single token
	if !l.Allow() {
		t.Fatal("first allow should succeed")
	}
	if l.Allow() {
		t.Fatal("second allow should fail immediately")
	}

	// Advance internal clock by injecting via the exported now field isn't
	// possible here; instead we sleep briefly to let tokens refill.
	_ = now
	time.Sleep(150 * time.Millisecond) // 10 tok/s → 1 token in 100ms
	if !l.Allow() {
		t.Fatal("expected allow after refill period")
	}
}

func TestDefaultOptions_SaneValues(t *testing.T) {
	opts := ratelimit.DefaultOptions()
	if opts.Rate <= 0 {
		t.Errorf("expected positive rate, got %v", opts.Rate)
	}
	if opts.Burst <= 0 {
		t.Errorf("expected positive burst, got %v", opts.Burst)
	}
}

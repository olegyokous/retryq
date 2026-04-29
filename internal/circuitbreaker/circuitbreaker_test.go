package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/andygeiss/retryq/internal/circuitbreaker"
)

func opts(max int, cooldown time.Duration) circuitbreaker.Options {
	return circuitbreaker.Options{MaxFailures: max, Cooldown: cooldown}
}

func TestAllow_ClosedByDefault(t *testing.T) {
	b := circuitbreaker.New(opts(3, time.Second))
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordFailure_OpensAfterThreshold(t *testing.T) {
	b := circuitbreaker.New(opts(3, time.Second))
	b.RecordFailure()
	b.RecordFailure()
	if b.State() != circuitbreaker.StateClosed {
		t.Fatal("expected closed before threshold")
	}
	b.RecordFailure()
	if b.State() != circuitbreaker.StateOpen {
		t.Fatal("expected open after threshold")
	}
}

func TestAllow_RejectsWhenOpen(t *testing.T) {
	b := circuitbreaker.New(opts(1, time.Hour))
	b.RecordFailure()
	if err := b.Allow(); err != circuitbreaker.ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestAllow_ProbeAfterCooldown(t *testing.T) {
	b := circuitbreaker.New(opts(1, time.Millisecond))
	b.RecordFailure()
	time.Sleep(5 * time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected probe to be allowed after cooldown, got %v", err)
	}
	if b.State() != circuitbreaker.StateClosed {
		t.Fatal("expected state to be closed after probe")
	}
}

func TestRecordSuccess_ResetFailures(t *testing.T) {
	b := circuitbreaker.New(opts(3, time.Second))
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	// Should need 3 more failures to open.
	b.RecordFailure()
	b.RecordFailure()
	if b.State() != circuitbreaker.StateClosed {
		t.Fatal("expected closed after success reset")
	}
	b.RecordFailure()
	if b.State() != circuitbreaker.StateOpen {
		t.Fatal("expected open after 3 failures post-reset")
	}
}

func TestDefaultOptions_SaneValues(t *testing.T) {
	o := circuitbreaker.DefaultOptions()
	if o.MaxFailures <= 0 {
		t.Errorf("MaxFailures should be positive, got %d", o.MaxFailures)
	}
	if o.Cooldown <= 0 {
		t.Errorf("Cooldown should be positive, got %v", o.Cooldown)
	}
}

func TestNew_ZeroOptions_UsesDefaults(t *testing.T) {
	b := circuitbreaker.New(circuitbreaker.Options{})
	if b == nil {
		t.Fatal("expected non-nil breaker")
	}
	if err := b.Allow(); err != nil {
		t.Fatalf("expected allow on fresh breaker, got %v", err)
	}
}

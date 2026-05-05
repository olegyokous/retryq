package jitter_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/sashajdn/retryq/internal/jitter"
)

func deterministicApplier(strategy jitter.Strategy) *jitter.Applier {
	//nolint:gosec
	rng := rand.New(rand.NewSource(42))
	return jitter.New(jitter.Options{Strategy: strategy, Rand: rng})
}

func TestApply_None_ReturnsSameDuration(t *testing.T) {
	a := deterministicApplier(jitter.None)
	base := 500 * time.Millisecond
	if got := a.Apply(base); got != base {
		t.Fatalf("None strategy: want %v, got %v", base, got)
	}
}

func TestApply_Full_WithinRange(t *testing.T) {
	a := deterministicApplier(jitter.Full)
	base := 1 * time.Second
	for i := 0; i < 200; i++ {
		got := a.Apply(base)
		if got < 0 || got > base {
			t.Fatalf("Full strategy out of range [0, %v]: got %v", base, got)
		}
	}
}

func TestApply_Equal_WithinRange(t *testing.T) {
	a := deterministicApplier(jitter.Equal)
	base := 1 * time.Second
	half := base / 2
	for i := 0; i < 200; i++ {
		got := a.Apply(base)
		if got < half || got > base {
			t.Fatalf("Equal strategy out of range [%v, %v]: got %v", half, base, got)
		}
	}
}

func TestApply_ZeroBase_ReturnsZero(t *testing.T) {
	a := deterministicApplier(jitter.Full)
	if got := a.Apply(0); got != 0 {
		t.Fatalf("expected 0 for zero base, got %v", got)
	}
}

func TestApply_NegativeBase_ReturnsUnchanged(t *testing.T) {
	a := deterministicApplier(jitter.Full)
	base := -1 * time.Second
	if got := a.Apply(base); got != base {
		t.Fatalf("expected negative base unchanged, got %v", got)
	}
}

func TestDefaultOptions_StrategyIsFull(t *testing.T) {
	opts := jitter.DefaultOptions()
	if opts.Strategy != jitter.Full {
		t.Fatalf("expected Full strategy, got %v", opts.Strategy)
	}
}

func TestNew_NilRand_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	a := jitter.New(jitter.DefaultOptions())
	_ = a.Apply(100 * time.Millisecond)
}

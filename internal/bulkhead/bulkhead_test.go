package bulkhead_test

import (
	"sync"
	"testing"

	"github.com/andygeiss/retryq/internal/bulkhead"
)

func TestAcquire_AllowsUnderLimit(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 3})
	if err := b.Acquire(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	b.Release()
}

func TestAcquire_RejectsWhenFull(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 2})
	_ = b.Acquire()
	_ = b.Acquire()
	if err := b.Acquire(); err != bulkhead.ErrFull {
		t.Fatalf("expected ErrFull, got %v", err)
	}
	b.Release()
	b.Release()
}

func TestRelease_FreesSlot(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 1})
	_ = b.Acquire()
	b.Release()
	if err := b.Acquire(); err != nil {
		t.Fatalf("slot should be free after release, got %v", err)
	}
	b.Release()
}

func TestInflight_TracksCorrectly(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 10})
	_ = b.Acquire()
	_ = b.Acquire()
	if got := b.Inflight(); got != 2 {
		t.Fatalf("expected 2 inflight, got %d", got)
	}
	b.Release()
	if got := b.Inflight(); got != 1 {
		t.Fatalf("expected 1 inflight, got %d", got)
	}
	b.Release()
}

func TestAcquire_ZeroMax_AlwaysAllows(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 0})
	for i := 0; i < 1000; i++ {
		if err := b.Acquire(); err != nil {
			t.Fatalf("expected nil for disabled bulkhead, got %v", err)
		}
	}
}

func TestAcquire_ConcurrentSafety(t *testing.T) {
	b := bulkhead.New(bulkhead.Options{MaxConcurrent: 50})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.Acquire(); err == nil {
				defer b.Release()
			}
		}()
	}
	wg.Wait()
	if got := b.Inflight(); got != 0 {
		t.Fatalf("expected 0 inflight after all released, got %d", got)
	}
}

func TestDefaultOptions_SaneValues(t *testing.T) {
	opts := bulkhead.DefaultOptions()
	if opts.MaxConcurrent <= 0 {
		t.Fatalf("expected positive MaxConcurrent, got %d", opts.MaxConcurrent)
	}
}

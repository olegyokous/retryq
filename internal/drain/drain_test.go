package drain_test

import (
	"context"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/retryq/internal/drain"
)

// stubWaiter implements drain.Waiter via an atomic counter so tests can
// simulate items completing over time.
type stubWaiter struct{ n atomic.Int32 }

func (s *stubWaiter) Inflight() int { return int(s.n.Load()) }
func (s *stubWaiter) set(n int)     { s.n.Store(int32(n)) }

var silentLogger = log.New(os.Discard, "", 0)

func silentOpts() drain.Options {
	o := drain.DefaultOptions()
	o.Logger = silentLogger
	o.PollInterval = 5 * time.Millisecond
	return o
}

func TestWait_DrainedImmediately(t *testing.T) {
	w := &stubWaiter{}
	ctx := context.Background()
	ok := drain.Wait(ctx, w, silentOpts())
	if !ok {
		t.Fatal("expected true when queue is already empty")
	}
}

func TestWait_DrainedAfterDelay(t *testing.T) {
	w := &stubWaiter{}
	w.set(2)
	go func() {
		time.Sleep(20 * time.Millisecond)
		w.set(0)
	}()
	opts := silentOpts()
	opts.Timeout = 500 * time.Millisecond
	ok := drain.Wait(context.Background(), w, opts)
	if !ok {
		t.Fatal("expected true when items finish before timeout")
	}
}

func TestWait_TimeoutExceeded(t *testing.T) {
	w := &stubWaiter{}
	w.set(5)
	opts := silentOpts()
	opts.Timeout = 30 * time.Millisecond
	ok := drain.Wait(context.Background(), w, opts)
	if ok {
		t.Fatal("expected false when timeout is exceeded")
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	w := &stubWaiter{}
	w.set(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	opts := silentOpts()
	opts.Timeout = 10 * time.Second
	ok := drain.Wait(ctx, w, opts)
	if ok {
		t.Fatal("expected false when context is cancelled")
	}
}

func TestDefaultOptions_SaneValues(t *testing.T) {
	o := drain.DefaultOptions()
	if o.Timeout < time.Second {
		t.Errorf("timeout too short: %v", o.Timeout)
	}
	if o.PollInterval <= 0 {
		t.Errorf("poll interval must be positive, got %v", o.PollInterval)
	}
}

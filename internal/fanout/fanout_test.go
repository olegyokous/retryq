package fanout_test

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/retryq/internal/fanout"
)

// stubDispatcher records calls and optionally returns an error for a target.
type stubDispatcher struct {
	calls   atomic.Int64
	errMap  map[string]error
}

func (s *stubDispatcher) Dispatch(_ context.Context, target string, _ http.Header, _ []byte) error {
	s.calls.Add(1)
	if s.errMap != nil {
		return s.errMap[target]
	}
	return nil
}

func TestFan_AllSucceed(t *testing.T) {
	d := &stubDispatcher{}
	targets := []string{"http://a.example", "http://b.example", "http://c.example"}

	results := fanout.Fan(context.Background(), d, targets, nil, []byte(`{}`), fanout.DefaultOptions())

	if got := len(results); got != 3 {
		t.Fatalf("expected 3 results, got %d", got)
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.TargetURL, r.Err)
		}
	}
	if d.calls.Load() != 3 {
		t.Errorf("expected 3 dispatch calls, got %d", d.calls.Load())
	}
}

func TestFan_PartialFailure(t *testing.T) {
	sentinel := errors.New("dispatch failed")
	d := &stubDispatcher{
		errMap: map[string]error{
			"http://bad.example": sentinel,
		},
	}
	targets := []string{"http://ok.example", "http://bad.example"}

	results := fanout.Fan(context.Background(), d, targets, nil, nil, fanout.DefaultOptions())

	if results[0].Err != nil {
		t.Errorf("expected ok target to succeed")
	}
	if !errors.Is(results[1].Err, sentinel) {
		t.Errorf("expected sentinel error, got %v", results[1].Err)
	}
}

func TestFan_EmptyTargets_ReturnsNil(t *testing.T) {
	d := &stubDispatcher{}
	results := fanout.Fan(context.Background(), d, nil, nil, nil, fanout.DefaultOptions())
	if results != nil {
		t.Errorf("expected nil results for empty targets")
	}
}

func TestFan_PreservesOrder(t *testing.T) {
	d := &stubDispatcher{}
	targets := []string{"http://one", "http://two", "http://three"}

	results := fanout.Fan(context.Background(), d, targets, nil, nil, fanout.DefaultOptions())

	for i, r := range results {
		if r.TargetURL != targets[i] {
			t.Errorf("index %d: expected %s, got %s", i, targets[i], r.TargetURL)
		}
	}
}

func TestFan_RespectsTimeout(t *testing.T) {
	slow := &slowDispatcher{delay: 200 * time.Millisecond}
	opts := fanout.Options{Timeout: 50 * time.Millisecond}

	start := time.Now()
	results := fanout.Fan(context.Background(), slow, []string{"http://slow"}, nil, nil, opts)
	elapsed := time.Since(start)

	if elapsed > 150*time.Millisecond {
		t.Errorf("fan took too long (%v); timeout not applied", elapsed)
	}
	if results[0].Err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestFirstError_AllOK(t *testing.T) {
	results := []fanout.Result{{TargetURL: "http://a"}, {TargetURL: "http://b"}}
	if err := fanout.FirstError(results); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestFirstError_ReturnsFirstErr(t *testing.T) {
	sentinel := errors.New("boom")
	results := []fanout.Result{
		{TargetURL: "http://a"},
		{TargetURL: "http://b", Err: sentinel},
	}
	if err := fanout.FirstError(results); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// slowDispatcher sleeps for delay then returns a context error.
type slowDispatcher struct{ delay time.Duration }

func (s *slowDispatcher) Dispatch(ctx context.Context, _ string, _ http.Header, _ []byte) error {
	select {
	case <-time.After(s.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

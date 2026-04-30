package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/retryq/internal/queue"
	"github.com/example/retryq/internal/retry"
)

// stubEnqueuer records calls to RecordFailure.
type stubEnqueuer struct {
	called int
	err    error
}

func (s *stubEnqueuer) RecordFailure(_ *queue.Item) error {
	s.s.called++
	return s.err
}

// fakePolicy always returns a fixed delay.
type fakePolicy struct{ d time.Duration }

func (f fakePolicy) Next(_ int) time.Duration { return f.d }

func item() *queue.Item {
	return &queue.Item{ID: "abc", Attempts: 1, TargetURL: "http://example.com"}
}

func TestSchedule_CallsRecordFailure(t *testing.T) {
	enq := &stubEnqueuer{}
	s := retry.New(enq, fakePolicy{d: 2 * time.Second}, nil, nil)

	if err := s.Schedule(context.Background(), item()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enq.called != 1 {
		t.Fatalf("expected 1 RecordFailure call, got %d", enq.called)
	}
}

func TestSchedule_PropagatesEnqueueError(t *testing.T) {
	want := errors.New("queue full")
	enq := &stubEnqueuer{err: want}
	s := retry.New(enq, fakePolicy{d: time.Second}, nil, nil)

	got := s.Schedule(context.Background(), item())
	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestSchedule_NilLogger_DoesNotPanic(t *testing.T) {
	enq := &stubEnqueuer{}
	s := retry.New(enq, fakePolicy{d: time.Second}, nil, nil)
	// Should not panic even without an explicit logger.
	_ = s.Schedule(context.Background(), item())
}

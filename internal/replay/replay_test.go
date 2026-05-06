package replay_test

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/your-org/retryq/internal/deadletter"
	"github.com/your-org/retryq/internal/replay"
)

var silentLog = log.New(io.Discard, "", 0)

type stubDispatcher struct {
	err error
	calls int
}

func (s *stubDispatcher) Dispatch(_ context.Context, _ deadletter.Entry) error {
	s.calls++
	return s.err
}

func makeEntry(id string) deadletter.Entry {
	return deadletter.Entry{
		ID:        id,
		TargetURL: "http://example.com",
		Payload:   []byte(`{}`),
		CreatedAt: time.Now(),
	}
}

func TestReplay_AllSucceed(t *testing.T) {
	d := &stubDispatcher{}
	entries := []deadletter.Entry{makeEntry("a"), makeEntry("b"), makeEntry("c")}
	opts := replay.DefaultOptions()
	opts.Logger = silentLog

	results := replay.Replay(context.Background(), entries, d, opts)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Entry.ID, r.Err)
		}
	}
}

func TestReplay_PartialFailure(t *testing.T) {
	sentinel := errors.New("dispatch error")
	d := &stubDispatcher{err: sentinel}
	entries := []deadletter.Entry{makeEntry("x")}
	opts := replay.DefaultOptions()
	opts.Logger = silentLog

	results := replay.Replay(context.Background(), entries, d, opts)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !errors.Is(results[0].Err, sentinel) {
		t.Errorf("expected sentinel error, got %v", results[0].Err)
	}
}

func TestReplay_EmptyEntries(t *testing.T) {
	d := &stubDispatcher{}
	opts := replay.DefaultOptions()
	opts.Logger = silentLog

	results := replay.Replay(context.Background(), nil, d, opts)

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
	if d.calls != 0 {
		t.Errorf("expected 0 dispatch calls, got %d", d.calls)
	}
}

func TestReplay_ConcurrencyDefault(t *testing.T) {
	d := &stubDispatcher{}
	opts := replay.Options{Concurrency: 0, Logger: silentLog} // 0 → defaults to 1
	entries := []deadletter.Entry{makeEntry("1"), makeEntry("2")}

	results := replay.Replay(context.Background(), entries, d, opts)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

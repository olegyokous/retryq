package audit_test

import (
	"testing"
	"time"

	"github.com/sablierapp/retryq/internal/audit"
)

func TestRecord_AddsEvent(t *testing.T) {
	l := audit.New(10)
	l.Record(audit.Event{
		ID:        "abc",
		Kind:      audit.EventEnqueued,
		TargetURL: "http://example.com",
	})
	if l.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", l.Len())
	}
}

func TestRecord_SetsTimestampIfZero(t *testing.T) {
	l := audit.New(10)
	before := time.Now().UTC()
	l.Record(audit.Event{ID: "t1", Kind: audit.EventDispatched})
	events := l.List()
	if events[0].Timestamp.Before(before) {
		t.Error("expected timestamp to be set automatically")
	}
}

func TestRecord_EvictsOldestWhenFull(t *testing.T) {
	l := audit.New(3)
	for i, id := range []string{"a", "b", "c", "d"} {
		l.Record(audit.Event{ID: id, Kind: audit.EventEnqueued, Attempt: i})
	}
	events := l.List()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].ID != "b" {
		t.Errorf("expected oldest evicted, got first ID %q", events[0].ID)
	}
}

func TestList_ReturnsCopy(t *testing.T) {
	l := audit.New(10)
	l.Record(audit.Event{ID: "x", Kind: audit.EventRetry})
	a := l.List()
	a[0].ID = "mutated"
	b := l.List()
	if b[0].ID == "mutated" {
		t.Error("List should return a copy, not a reference")
	}
}

func TestNew_DefaultMaxSize(t *testing.T) {
	l := audit.New(0)
	for i := 0; i < 1001; i++ {
		l.Record(audit.Event{Kind: audit.EventEnqueued})
	}
	if l.Len() != audit.DefaultMaxSize {
		t.Errorf("expected %d, got %d", audit.DefaultMaxSize, l.Len())
	}
}

func TestLen_ReflectsSize(t *testing.T) {
	l := audit.New(10)
	if l.Len() != 0 {
		t.Error("expected empty log")
	}
	l.Record(audit.Event{Kind: audit.EventDeadLetter})
	l.Record(audit.Event{Kind: audit.EventRequeued})
	if l.Len() != 2 {
		t.Errorf("expected 2, got %d", l.Len())
	}
}

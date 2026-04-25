package deadletter

import (
	"testing"
)

func TestPop_RemovesAndReturnsEntry(t *testing.T) {
	s := New(10)
	e := newEntry("abc", "https://example.com")
	s.Add(e)

	got, err := s.Pop(e.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != e.ID {
		t.Errorf("expected ID %q, got %q", e.ID, got.ID)
	}
	if s.Len() != 0 {
		t.Errorf("expected store to be empty after pop, got len %d", s.Len())
	}
}

func TestPop_NotFound(t *testing.T) {
	s := New(10)

	_, err := s.Pop("nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPop_OnlyRemovesTargetEntry(t *testing.T) {
	s := New(10)
	e1 := newEntry("id-1", "https://a.example.com")
	e2 := newEntry("id-2", "https://b.example.com")
	s.Add(e1)
	s.Add(e2)

	_, err := s.Pop(e1.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Len() != 1 {
		t.Errorf("expected 1 remaining entry, got %d", s.Len())
	}
	list := s.List()
	if list[0].ID != e2.ID {
		t.Errorf("expected remaining entry to be %q, got %q", e2.ID, list[0].ID)
	}
}

func TestLen_ReflectsStoreSize(t *testing.T) {
	s := New(10)
	if s.Len() != 0 {
		t.Errorf("expected 0, got %d", s.Len())
	}
	s.Add(newEntry("x", "https://x.example.com"))
	if s.Len() != 1 {
		t.Errorf("expected 1, got %d", s.Len())
	}
}

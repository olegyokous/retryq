// Package priority provides a simple priority level type and helpers for
// classifying retry queue items as high, normal, or low priority so that
// the worker can drain higher-priority items first.
package priority

import (
	"errors"
	"strings"
)

// Level represents the dispatch priority of a queue item.
type Level int

const (
	// Low priority items are processed last.
	Low Level = iota
	// Normal is the default priority level.
	Normal
	// High priority items are processed first.
	High
)

// String returns a human-readable name for the priority level.
func (l Level) String() string {
	switch l {
	case High:
		return "high"
	case Normal:
		return "normal"
	case Low:
		return "low"
	default:
		return "unknown"
	}
}

// Parse converts a case-insensitive string ("high", "normal", "low") into a
// Level. It returns Normal and a nil error for an empty string so that callers
// that omit the field get a sensible default.
func Parse(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "normal":
		return Normal, nil
	case "high":
		return High, nil
	case "low":
		return Low, nil
	default:
		return Normal, errors.New("priority: unknown level " + s)
	}
}

// MustParse is like Parse but panics on an unrecognised value. Intended for
// use in tests and static initialisers.
func MustParse(s string) Level {
	l, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return l
}

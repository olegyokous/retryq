// Package tags provides structured key-value tagging for retry queue items,
// enabling filtering, routing, and observability by arbitrary metadata.
package tags

import (
	"fmt"
	"strings"
)

// MaxTags is the maximum number of tags allowed per item.
const MaxTags = 16

// ErrTooManyTags is returned when adding a tag would exceed MaxTags.
var ErrTooManyTags = fmt.Errorf("tags: cannot exceed %d tags per item", MaxTags)

// Tags is an immutable snapshot of key-value metadata attached to a queue item.
type Tags map[string]string

// Builder accumulates tags before attaching them to an item.
type Builder struct {
	tags map[string]string
}

// New returns an empty Builder.
func New() *Builder {
	return &Builder{tags: make(map[string]string)}
}

// Set adds or replaces a tag. Returns ErrTooManyTags if the limit would be exceeded.
func (b *Builder) Set(key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("tags: key must not be empty")
	}
	if _, exists := b.tags[key]; !exists && len(b.tags) >= MaxTags {
		return ErrTooManyTags
	}
	b.tags[key] = value
	return nil
}

// MustSet calls Set and panics on error. Useful for static initialisation.
func (b *Builder) MustSet(key, value string) *Builder {
	if err := b.Set(key, value); err != nil {
		panic(err)
	}
	return b
}

// Build returns an immutable Tags snapshot.
func (b *Builder) Build() Tags {
	out := make(Tags, len(b.tags))
	for k, v := range b.tags {
		out[k] = v
	}
	return out
}

// Get returns the value for key and whether it was present.
func (t Tags) Get(key string) (string, bool) {
	v, ok := t[key]
	return v, ok
}

// Matches reports whether all entries in filter are present in t with equal values.
func (t Tags) Matches(filter Tags) bool {
	for k, v := range filter {
		if t[k] != v {
			return false
		}
	}
	return true
}

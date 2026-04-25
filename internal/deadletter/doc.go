// Package deadletter provides an in-memory dead-letter store for HTTP
// requests that have exhausted all retry attempts.
//
// The store holds a bounded ring-buffer of [Entry] values. When the store
// is full the oldest entry is evicted to make room for the new one, so
// memory usage is always capped at the configured maximum size.
//
// Typical usage:
//
//	store := deadletter.New(deadletter.WithMaxSize(500))
//
//	// record a failed item
//	store.Add(deadletter.Entry{
//		ID:          item.ID,
//		TargetURL:   item.TargetURL,
//		Payload:     item.Payload,
//		LastAttempt: time.Now(),
//		Reason:      err.Error(),
//	})
//
//	// inspect all dead-lettered entries
//	entries := store.List()
package deadletter

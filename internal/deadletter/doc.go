// Package deadletter provides storage and requeue functionality for
// HTTP requests that have exhausted all retry attempts.
//
// A Store holds a bounded ring-buffer of dead-letter entries. Each entry
// captures the original request payload along with metadata about why it
// failed. Entries can be listed for inspection or selectively requeued
// back into the retry pipeline via Pop + Enqueue.
//
// # Typical usage
//
//	store := deadletter.NewStore(100)
//
//	// After all retries are exhausted, persist the failed request:
//	store.Push(entry)
//
//	// Later, inspect accumulated failures:
//	for _, e := range store.List() {
//		log.Printf("dead letter: %s %s – last error: %v", e.Method, e.URL, e.LastErr)
//	}
//
//	// Requeue a specific entry back into the retry pipeline:
//	if e, ok := store.Pop(id); ok {
//		queue.Enqueue(e.Request)
//	}
package deadletter

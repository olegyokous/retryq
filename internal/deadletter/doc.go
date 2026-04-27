// Package deadletter provides storage and requeue functionality for
// HTTP requests that have exhausted all retry attempts.
//
// A Store holds a bounded ring-buffer of dead-letter entries. Each entry
// captures the original request payload along with metadata about why it
// failed. Entries can be listed for inspection or selectively requeued
// back into the retry pipeline via Pop + Enqueue.
package deadletter

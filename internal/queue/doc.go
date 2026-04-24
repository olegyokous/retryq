// Package queue provides an in-memory retry queue for HTTP requests.
//
// Items are enqueued with a target URL and HTTP payload. The queue
// tracks retry attempts and uses an exponential backoff policy to
// schedule subsequent delivery attempts. Items that exceed the
// configured maximum attempts are moved to the dead-letter bucket
// where they can be inspected or re-queued by the caller.
//
// Basic usage:
//
//	cfg := config.Default()
//	q := queue.New(cfg)
//
//	q.Enqueue(&queue.Item{
//		ID:     "evt-123",
//		Method: "POST",
//		URL:    "https://example.com/webhook",
//		Body:   payload,
//	})
//
//	for _, item := range q.Ready(ctx) {
//		if err := deliver(item); err != nil {
//			q.RecordFailure(item.ID, err.Error())
//		} else {
//			q.RecordSuccess(item.ID)
//		}
//	}
package queue

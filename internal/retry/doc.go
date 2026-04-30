// Package retry wires together the backoff policy and the queue to
// provide a single Schedule entry-point used by the worker layer.
//
// Typical usage:
//
//	pol := backoff.New(cfg)
//	sched := retry.New(q, pol, m, logger)
//	// inside worker on dispatch failure:
//	if err := sched.Schedule(ctx, item); err != nil {
//	    log.Error("could not schedule retry", "err", err)
//	}
//
// The Scheduler is intentionally thin: it delegates persistence to the
// queue and delay calculation to the backoff policy, making each layer
// independently testable.
package retry

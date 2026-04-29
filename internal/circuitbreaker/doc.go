// Package circuitbreaker implements a thread-safe, per-target circuit breaker
// for use with the retryq worker.
//
// # Overview
//
// A Breaker tracks consecutive failures for a given target. Once the failure
// count reaches MaxFailures the circuit transitions to the open state and
// subsequent calls to Allow return ErrOpen, preventing further dispatch
// attempts until the Cooldown period elapses.
//
// After the cooldown a single probe request is permitted (half-open). If the
// probe succeeds (RecordSuccess) the circuit closes and normal operation
// resumes. If it fails (RecordFailure) the circuit opens again.
//
// # Usage
//
//	breaker := circuitbreaker.New(circuitbreaker.DefaultOptions())
//
//	if err := breaker.Allow(); err != nil {
//		// circuit is open — skip dispatch
//		return err
//	}
//
//	if err := dispatch(item); err != nil {
//		breaker.RecordFailure()
//		return err
//	}
//
//	breaker.RecordSuccess()
package circuitbreaker

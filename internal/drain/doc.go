// Package drain implements graceful-shutdown helpers for retryq.
//
// # Overview
//
// When a SIGTERM is received the process should stop accepting new work
// and wait for any in-flight HTTP dispatches to complete before exiting.
// This package provides two complementary tools:
//
//   - [Wait] blocks the calling goroutine (typically inside main) until
//     the [Waiter]'s inflight count reaches zero or a deadline is hit.
//
//   - [NewHandler] exposes a lightweight HTTP endpoint (/drain) that
//     external orchestrators (Kubernetes readiness probes, load-balancer
//     health checks) can poll to confirm the instance has drained.
//
// # Typical usage
//
//	sigCh := make(chan os.Signal, 1)
//	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
//	<-sigCh
//
//	opts := drain.DefaultOptions()
//	if ok := drain.Wait(ctx, bulkhead, opts); !ok {
//	    log.Println("drain: shutdown forced")
//	}
package drain

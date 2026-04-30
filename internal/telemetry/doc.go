// Package telemetry provides lightweight distributed tracing primitives for
// retryq.
//
// # Overview
//
// Every HTTP request entering retryq is assigned a TraceID — a 32-character
// hex-encoded random string. The ID is:
//
//  1. Read from the incoming X-Retryq-Trace-Id header if present, so callers
//     can propagate their own trace context.
//  2. Generated fresh when the header is absent.
//  3. Stored in the request [context.Context] for the lifetime of the request.
//  4. Echoed back in the response X-Retryq-Trace-Id header.
//
// # Usage
//
// Wrap your HTTP mux with [Middleware] during server initialisation:
//
//	handler = telemetry.Middleware(handler)
//
// Retrieve the ID anywhere you have access to the context:
//
//	id, ok := telemetry.FromContext(ctx)
//
// Attach it to structured log lines with [LogAttr]:
//
//	slog.InfoContext(ctx, "item enqueued", telemetry.LogAttr(id))
package telemetry

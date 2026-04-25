// Package server provides the HTTP ingestion layer for retryq.
//
// It exposes two endpoints:
//
//	POST /enqueue  – accepts a JSON payload describing the outbound
//	                 HTTP request to be attempted and places it on the
//	                 retry queue.  Returns 202 Accepted with the
//	                 generated item ID on success.
//
//	GET  /healthz  – liveness probe; always returns 200 OK.
//
// Example request body for /enqueue:
//
//	{
//	  "target_url": "https://api.example.com/webhook",
//	  "method":     "POST",
//	  "headers":    {"Authorization": "Bearer token"},
//	  "body":       "{\"event\":\"order.created\"}"
//	}
package server

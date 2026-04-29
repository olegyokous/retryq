// Package main is the entry point for the retryq HTTP retry queue service.
// It wires together all internal components and starts the HTTP server.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/retryq/internal/backoff"
	"github.com/yourorg/retryq/internal/config"
	"github.com/yourorg/retryq/internal/deadletter"
	"github.com/yourorg/retryq/internal/dispatcher"
	"github.com/yourorg/retryq/internal/metrics"
	"github.com/yourorg/retryq/internal/queue"
	"github.com/yourorg/retryq/internal/server"
	"github.com/yourorg/retryq/internal/worker"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	workers := flag.Int("workers", 4, "number of concurrent dispatch workers")
	maxDLSize := flag.Int("dl-max-size", 1000, "maximum dead-letter store capacity")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Build retry policy from defaults.
	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	// Initialise Prometheus metrics on the default registry.
	reg := prometheus.DefaultRegisterer
	m := metrics.New(reg)

	// Build the backoff policy.
	policy := backoff.New(cfg)

	// Build the retry queue.
	q := queue.New(cfg, m)

	// Build the dead-letter store.
	dlStore := deadletter.New(deadletter.WithMaxSize(*maxDLSize))

	// Build the HTTP dispatcher.
	disp := dispatcher.New(
		dispatcher.WithTimeout(10*time.Second),
		dispatcher.WithClient(&http.Client{}),
	)

	// Start worker pool.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < *workers; i++ {
		w := worker.New(worker.Deps{
			Queue:      q,
			Dispatcher: disp,
			Backoff:    policy,
			DeadLetter: dlStore,
			Metrics:    m,
			Logger:     slog.Default(),
		})
		go w.Run(ctx)
	}

	slog.Info("worker pool started", "count", *workers)

	// Build and start the HTTP server.
	srv := server.New(server.Deps{
		Queue:      q,
		DeadLetter: dlStore,
		Metrics:    m,
		Logger:     slog.Default(),
	})

	httpServer := &http.Server{
		Addr:         *addr,
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", *addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}

	slog.Info("shutdown complete")
}

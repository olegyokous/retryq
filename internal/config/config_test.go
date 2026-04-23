package config_test

import (
	"testing"
	"time"

	"github.com/yourorg/retryq/internal/config"
)

func TestDefault_IsValid(t *testing.T) {
	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid, got: %v", err)
	}
}

func TestValidate_MaxAttempts(t *testing.T) {
	cfg := config.Default()
	cfg.Retry.MaxAttempts = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for max_attempts=0")
	}
}

func TestValidate_InitialInterval(t *testing.T) {
	cfg := config.Default()
	cfg.Retry.InitialInterval = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for initial_interval=0")
	}
}

func TestValidate_MultiplierBelowOne(t *testing.T) {
	cfg := config.Default()
	cfg.Retry.Multiplier = 0.5
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for multiplier < 1.0")
	}
}

func TestValidate_MaxIntervalLessThanInitial(t *testing.T) {
	cfg := config.Default()
	cfg.Retry.InitialInterval = 5 * time.Second
	cfg.Retry.MaxInterval = 1 * time.Second
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when max_interval < initial_interval")
	}
}

func TestValidate_JitterOutOfRange(t *testing.T) {
	cfg := config.Default()
	cfg.Retry.Jitter = 1.5
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for jitter > 1.0")
	}
}

func TestValidate_Workers(t *testing.T) {
	cfg := config.Default()
	cfg.Workers = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for workers=0")
	}
}

func TestValidate_QueueCapacity(t *testing.T) {
	cfg := config.Default()
	cfg.QueueCapacity = -1
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for queue_capacity=-1")
	}
}

func TestDefault_Fields(t *testing.T) {
	cfg := config.Default()
	if cfg.Retry.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", cfg.Retry.MaxAttempts)
	}
	if cfg.Retry.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", cfg.Retry.Multiplier)
	}
	if cfg.Workers != 4 {
		t.Errorf("expected Workers=4, got %d", cfg.Workers)
	}
}

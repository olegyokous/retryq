package sampling_test

import (
	"testing"

	"github.com/yourorg/retryq/internal/sampling"
)

func TestSample_AlwaysAllowsAtRateOne(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 1.0, Seed: 1})
	for i := 0; i < 1000; i++ {
		if !s.Sample() {
			t.Fatalf("expected sample=true at rate 1.0, iteration %d", i)
		}
	}
}

func TestSample_NeverAllowsAtRateZero(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 0.0, Seed: 1})
	for i := 0; i < 1000; i++ {
		if s.Sample() {
			t.Fatalf("expected sample=false at rate 0.0, iteration %d", i)
		}
	}
}

func TestSample_ApproximateRateHalf(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 0.5, Seed: 99})
	hits := 0
	const n = 10_000
	for i := 0; i < n; i++ {
		if s.Sample() {
			hits++
		}
	}
	ratio := float64(hits) / n
	if ratio < 0.45 || ratio > 0.55 {
		t.Errorf("expected ratio near 0.5, got %.3f", ratio)
	}
}

func TestSample_ClampsNegativeRate(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: -5.0})
	if s.Rate() != 0 {
		t.Errorf("expected rate=0, got %f", s.Rate())
	}
}

func TestSample_ClampsRateAboveOne(t *testing.T) {
	s := sampling.New(sampling.Options{Rate: 3.5})
	if s.Rate() != 1.0 {
		t.Errorf("expected rate=1.0, got %f", s.Rate())
	}
}

func TestSample_NilSamplerAlwaysSamples(t *testing.T) {
	var s *sampling.Sampler
	if !s.Sample() {
		t.Error("nil Sampler.Sample() should return true")
	}
	if s.Rate() != 1.0 {
		t.Errorf("nil Sampler.Rate() should return 1.0, got %f", s.Rate())
	}
}

func TestDefaultOptions_RateIsOne(t *testing.T) {
	opts := sampling.DefaultOptions()
	if opts.Rate != 1.0 {
		t.Errorf("expected default rate=1.0, got %f", opts.Rate)
	}
}

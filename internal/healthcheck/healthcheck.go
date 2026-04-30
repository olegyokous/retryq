// Package healthcheck provides a structured health status reporter
// that aggregates the state of internal subsystems (queue, circuit breaker,
// dead-letter store) into a single JSON-serialisable response.
package healthcheck

import (
	"sync"
	"time"
)

// Status represents the health state of a single component.
type Status string

const (
	StatusOK      Status = "ok"
	StatusDegraded Status = "degraded"
	StatusDown    Status = "down"
)

// ComponentHealth holds the health information for one subsystem.
type ComponentHealth struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// Report is the top-level health response returned to callers.
type Report struct {
	Status     Status                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

// Checker is a named function that returns the health of a component.
type Checker struct {
	Name string
	Fn   func() ComponentHealth
}

// Aggregator collects multiple Checkers and produces a Report.
type Aggregator struct {
	mu       sync.RWMutex
	checkers []Checker
}

// New returns an empty Aggregator.
func New() *Aggregator {
	return &Aggregator{}
}

// Register adds a named health checker to the aggregator.
func (a *Aggregator) Register(c Checker) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkers = append(a.checkers, c)
}

// Check runs all registered checkers and returns an aggregated Report.
// The overall status is the worst status observed across all components.
func (a *Aggregator) Check() Report {
	a.mu.RLock()
	checkers := make([]Checker, len(a.checkers))
	copy(checkers, a.checkers)
	a.mu.RUnlock()

	components := make(map[string]ComponentHealth, len(checkers))
	overall := StatusOK

	for _, c := range checkers {
		h := c.Fn()
		components[c.Name] = h
		if worse(h.Status, overall) {
			overall = h.Status
		}
	}

	return Report{
		Status:     overall,
		Timestamp:  time.Now().UTC(),
		Components: components,
	}
}

// worse returns true when candidate is more severe than current.
func worse(candidate, current Status) bool {
	rank := map[Status]int{StatusOK: 0, StatusDegraded: 1, StatusDown: 2}
	return rank[candidate] > rank[current]
}

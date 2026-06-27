// Package breaker provides a circuit breaker pattern for protecting external calls.
// It supports three states: Closed, Open, and Half‑Open, with configurable failure
// threshold and reset timeout.
package breaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	Closed   State = iota // circuit is healthy, calls are allowed
	Open                  // circuit is tripped, calls are rejected immediately
	HalfOpen              // circuit is testing if the underlying service has recovered
)

func (s State) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// Breaker implements the circuit breaker pattern.
type Breaker struct {
	mu           sync.Mutex
	name         string
	state        State
	failures     int
	threshold    int           // number of consecutive failures to trip
	timeout      time.Duration // how long to stay Open before transitioning to Half‑Open
	lastFailTime time.Time
	nextTry      time.Time // when Half‑Open can be attempted again
}

// New creates a new circuit breaker.
//
//   - name: human‑readable identifier (for logging/metrics)
//   - threshold: number of consecutive failures before opening the circuit
//   - timeout: how long the breaker stays open before attempting a half‑open probe
func New(name string, threshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		name:      name,
		state:     Closed,
		threshold: threshold,
		timeout:   timeout,
	}
}

// ErrCircuitOpen is returned when the circuit is open and a call is rejected.
var ErrCircuitOpen = errors.New("breaker: circuit is open")

// Execute runs the given function if the circuit allows it.
// If the function returns an error, the failure is recorded.
// If the circuit is open, ErrCircuitOpen is returned without calling fn.
func (b *Breaker) Execute(ctx context.Context, fn func() error) error {
	if err := b.beforeCall(); err != nil {
		return err
	}
	err := fn()
	b.afterCall(err)
	return err
}

// beforeCall checks the circuit state and transitions if necessary.
func (b *Breaker) beforeCall() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case Closed:
		return nil
	case Open:
		if time.Now().After(b.nextTry) {
			b.state = HalfOpen
			return nil // allow one probe
		}
		return ErrCircuitOpen
	case HalfOpen:
		return nil // allow probe (only one goroutine at a time? Not enforced strictly, but after first success we close)
	}
	return nil
}

// afterCall updates the circuit state based on the outcome of the call.
func (b *Breaker) afterCall(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.failures++
		b.lastFailTime = time.Now()
		if b.failures >= b.threshold {
			b.state = Open
			b.nextTry = time.Now().Add(b.timeout)
		}
	} else {
		// success resets the circuit
		b.failures = 0
		b.state = Closed
	}
}

// State returns the current state of the breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Reset manually resets the breaker to the Closed state.
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = Closed
}

// Name returns the breaker's name.
func (b *Breaker) Name() string {
	return b.name
}

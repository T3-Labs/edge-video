package circuit

import (
	"fmt"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

type Breaker struct {
	name              string
	maxFailures       int64
	resetTimeout      time.Duration
	halfOpenSuccesses int

	// Backoff exponencial
	initialBackoff    time.Duration
	maxBackoff        time.Duration
	backoffMultiplier float64
	currentBackoff    time.Duration

	mu            sync.RWMutex
	state         State
	failures      int64
	successes     int64
	lastFailTime  time.Time
	lastStateTime time.Time
}

func NewBreaker(name string, maxFailures int64, resetTimeout time.Duration) *Breaker {
	// Use resetTimeout as initial backoff so unit tests with small timeouts work
	initialBackoff := resetTimeout

	return &Breaker{
		name:              name,
		maxFailures:       maxFailures,
		resetTimeout:      resetTimeout,
		halfOpenSuccesses: 3,
		initialBackoff:    initialBackoff,
		maxBackoff:        10 * time.Minute,
		backoffMultiplier: 2.0,
		currentBackoff:    initialBackoff,
		state:             StateClosed,
		lastStateTime:     time.Now(),
	}
}

func (cb *Breaker) Call(fn func() error) error {
	if !cb.Allow() {
		return fmt.Errorf("circuit breaker %s aberto", cb.name)
	}

	err := fn()

	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

func (cb *Breaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Use exponential backoff timing
		if time.Since(cb.lastFailTime) > cb.currentBackoff {
			cb.setState(StateHalfOpen)
			return true
		}
		return false

	case StateHalfOpen:
		return true

	default:
		return false
	}
}

func (cb *Breaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++

	switch cb.state {
	case StateClosed:
		cb.failures = 0
		// Reset backoff when returning to normal
		cb.currentBackoff = cb.initialBackoff

	case StateHalfOpen:
		if cb.successes >= int64(cb.halfOpenSuccesses) {
			cb.setState(StateClosed)
			cb.failures = 0
			cb.successes = 0
			// Reset backoff when returning to closed
			cb.currentBackoff = cb.initialBackoff
		}
	}
}

func (cb *Breaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()
	cb.successes = 0

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.setState(StateOpen)
			// Exponentially increase backoff when opening the circuit
			cb.currentBackoff = time.Duration(float64(cb.currentBackoff) * cb.backoffMultiplier)
			if cb.currentBackoff > cb.maxBackoff {
				cb.currentBackoff = cb.maxBackoff
			}
		}

	case StateHalfOpen:
		cb.setState(StateOpen)
		// Increase backoff when failing in half-open
		cb.currentBackoff = time.Duration(float64(cb.currentBackoff) * cb.backoffMultiplier)
		if cb.currentBackoff > cb.maxBackoff {
			cb.currentBackoff = cb.maxBackoff
		}
	}
}

func (cb *Breaker) setState(newState State) {
	if cb.state != newState {
		oldState := cb.state
		cb.state = newState
		cb.lastStateTime = time.Now()
		fmt.Printf("Circuit breaker %s: %s -> %s (falhas: %d, pr\u00f3xima tentativa em: %v)\n",
			cb.name, oldState, newState, cb.failures, cb.currentBackoff)
	}
}

func (cb *Breaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *Breaker) Stats() BreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return BreakerStats{
		Name:            cb.name,
		State:           cb.state,
		Failures:        cb.failures,
		Successes:       cb.successes,
		MaxFailures:     cb.maxFailures,
		CurrentBackoff:  cb.currentBackoff,
		LastFailTime:    cb.lastFailTime,
		LastStateChange: cb.lastStateTime,
	}
}

func (cb *Breaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.lastStateTime = time.Now()
}

type BreakerStats struct {
	Name            string
	State           State
	Failures        int64
	Successes       int64
	MaxFailures     int64
	CurrentBackoff  time.Duration
	LastFailTime    time.Time
	LastStateChange time.Time
}

func (bs BreakerStats) String() string {
	return fmt.Sprintf("Circuit[%s]: %s, Failures: %d/%d, Successes: %d, NextRetry: %v",
		bs.Name, bs.State, bs.Failures, bs.MaxFailures, bs.Successes, bs.CurrentBackoff)
}

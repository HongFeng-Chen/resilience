package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

func (s CircuitState) String() string {
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

var ErrCircuitOpen = errors.New("circuit breaker is open")

type OnBreakFunc func(err error, breakDuration time.Duration)
type OnResetFunc func()
type OnHalfOpenFunc func()

type CircuitBreaker struct {
	failureThreshold int
	breakDuration    time.Duration

	mutex sync.Mutex

	state CircuitState

	failures int

	lastFailureTime time.Time

	onBreak    OnBreakFunc
	onReset    OnResetFunc
	onHalfOpen OnHalfOpenFunc
}

// NewCircuitBreaker creates a circuit breaker policy
func NewCircuitBreaker(
	failureThreshold int,
	breakDuration time.Duration,
) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		breakDuration:    breakDuration,
		state:            Closed,
	}
}

func (c *CircuitBreaker) OnBreak(f OnBreakFunc) *CircuitBreaker {
	c.onBreak = f
	return c
}

func (c *CircuitBreaker) OnReset(f OnResetFunc) *CircuitBreaker {
	c.onReset = f
	return c
}

func (c *CircuitBreaker) OnHalfOpen(f OnHalfOpenFunc) *CircuitBreaker {
	c.onHalfOpen = f
	return c
}

func (c *CircuitBreaker) Execute(ctx context.Context, fn Func) error {
	// pre-check
	if err := c.beforeExecution(); err != nil {
		return err
	}

	err := fn(ctx)

	c.afterExecution(err)

	return err
}

func (c *CircuitBreaker) beforeExecution() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	switch c.state {
	case Open:
		if time.Since(c.lastFailureTime) >= c.breakDuration {
			c.transitionToHalfOpen()
			return nil
		}
		return ErrCircuitOpen

	case HalfOpen:
		// allow single trial
		return nil

	case Closed:
		return nil

	default:
		return nil
	}
}

func (c *CircuitBreaker) afterExecution(err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err == nil {
		if c.state == HalfOpen {
			c.reset()
		}
		return
	}

	// error occurred
	c.failures++

	if c.state == HalfOpen || c.failures >= c.failureThreshold {
		c.trip(err)
	}
}

func (c *CircuitBreaker) trip(err error) {
	c.state = Open
	c.lastFailureTime = time.Now()
	c.failures = 0

	if c.onBreak != nil {
		c.onBreak(err, c.breakDuration)
	}
}

func (c *CircuitBreaker) reset() {
	c.state = Closed
	c.failures = 0

	if c.onReset != nil {
		c.onReset()
	}
}

func (c *CircuitBreaker) transitionToHalfOpen() {
	c.state = HalfOpen

	if c.onHalfOpen != nil {
		c.onHalfOpen()
	}
}

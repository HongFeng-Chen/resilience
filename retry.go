package resilience

import (
	"context"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

/*
========================
 Retry Policy
========================
*/

type Retry struct {
	maxRetries   int
	retryForever bool

	shouldRetry func(error) bool
	backoff     BackoffStrategy
	onRetry     OnRetryFunc
}

// OnRetryFunc mirrors Polly's OnRetry callback
type OnRetryFunc func(
	attempt int,
	err error,
	delay time.Duration,
	ctx context.Context,
)

/*
========================
 Constructors
========================
*/

// New creates a retry policy with max retries
func NewRetry(maxRetries int) *Retry {
	return &Retry{
		maxRetries: maxRetries,
		shouldRetry: func(err error) bool {
			return err != nil
		},
		backoff: NoBackoff{},
	}
}

// Forever creates a retry-forever policy
func Forever() *Retry {
	return &Retry{
		retryForever: true,
		shouldRetry: func(err error) bool {
			return err != nil
		},
		backoff: NoBackoff{},
	}
}

/*
========================
 Fluent Configuration
========================
*/

// Handle configures retry condition
func (r *Retry) Handle(f func(error) bool) *Retry {
	r.shouldRetry = f
	return r
}

// WithBackoff configures backoff strategy
func (r *Retry) WithBackoff(b BackoffStrategy) *Retry {
	r.backoff = b
	return r
}

// OnRetry configures retry callback
func (r *Retry) OnRetry(f OnRetryFunc) *Retry {
	r.onRetry = f
	return r
}

// Execute retries based on error predicate
func (r *Retry) Execute(ctx context.Context, fn Func) error {
	var err error
	attempt := 0 // 记录重试次数

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = fn(ctx)
		if err == nil {
			return nil
		}

		if !r.shouldRetry(err) {
			return err
		}

		attempt++

		if !r.retryForever && attempt > r.maxRetries {
			return err
		}

		delay := r.backoff.Duration(attempt)

		if r.onRetry != nil {
			r.onRetry(attempt, err, delay, ctx)
		}

		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
}

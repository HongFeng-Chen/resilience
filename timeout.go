package resilience

import (
	"context"
	"errors"
	"time"
)

var ErrTimeout = errors.New("execution timed out")

type TimeoutMode int

const (
	// Optimistic uses context cancellation (recommended)
	Optimistic TimeoutMode = iota

	// Pessimistic enforces timeout regardless of fn behavior
	Pessimistic
)

type OnTimeoutFunc func(
	elapsed time.Duration, // 已过去的时间

	ctx context.Context, // 上下文环境

)

type Timeout struct {
	timeout   time.Duration
	mode      TimeoutMode
	onTimeout OnTimeoutFunc
}

func NewTimeout(timeout time.Duration) *Timeout {
	return &Timeout{
		timeout: timeout,
		mode:    Optimistic,
	}
}

func (t *Timeout) WithMode(mode TimeoutMode) *Timeout {
	t.mode = mode
	return t
}

func (t *Timeout) OnTimeout(fn OnTimeoutFunc) *Timeout {
	t.onTimeout = fn
	return t
}

func (t *Timeout) Execute(ctx context.Context, fn Func) error {
	switch t.mode {
	case Optimistic:
		return t.executeOptimistic(ctx, fn)
	case Pessimistic:
		return t.executePessimistic(ctx, fn)
	default:
		return t.executeOptimistic(ctx, fn)
	}
}

func (t *Timeout) executeOptimistic(ctx context.Context, fn Func) error {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	start := time.Now()
	err := fn(ctx)

	if errors.Is(err, context.DeadlineExceeded) {
		t.trigger(start, ctx)
		return ErrTimeout
	}

	if ctx.Err() == context.DeadlineExceeded {
		t.trigger(start, ctx)
		return ErrTimeout
	}

	return err
}

func (t *Timeout) executePessimistic(ctx context.Context, fn Func) error {
	result := make(chan error, 1)
	start := time.Now()

	go func() {
		result <- fn(ctx)
	}()

	select {
	case err := <-result:
		return err
	case <-time.After(t.timeout):
		t.trigger(start, ctx)
		return ErrTimeout
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *Timeout) trigger(start time.Time, ctx context.Context) {
	if t.onTimeout != nil {
		t.onTimeout(time.Since(start), ctx)
	}
}

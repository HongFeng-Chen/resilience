package resilience

import (
	"context"
	"errors"
	"sync"
)

var ErrBulkheadRejected = errors.New("bulkhead limit exceeded")

type OnBulkheadRejectedFunc func(ctx context.Context)

// Bulkhead implements the Resilience interface
type Bulkhead struct {
	maxParallel int // max concurrent executions
	maxQueue    int // max queued executions

	sem   chan struct{} // semaphore for concurrent executions
	queue chan struct{} // queue for queued executions

	onRejected OnBulkheadRejectedFunc // on bulkhead limit exceeded

	once sync.Once // for lazy initialization
}

// NewBulkhead creates a bulkhead policy
func NewBulkhead(maxParallel, maxQueue int) *Bulkhead {
	if maxParallel <= 0 {
		panic("maxParallel must be > 0")
	}
	if maxQueue < 0 {
		panic("maxQueue must be >= 0")
	}

	return &Bulkhead{
		maxParallel: maxParallel,
		maxQueue:    maxQueue,
	}
}

// OnRejected sets the callback for bulkhead rejections
// 当 Bulkhead 拒绝请求时的回调函数。
func (b *Bulkhead) OnRejected(fn OnBulkheadRejectedFunc) *Bulkhead {
	b.onRejected = fn
	return b
}

// recently viewed file: bulkhead/bulkhead.go
// Execute executes the given function with bulkhead policy
func (b *Bulkhead) Execute(ctx context.Context, fn Func) error {
	b.init()

	// Try enter execution slot immediately
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
		return fn(ctx)

	default:
	}

	// Try enter queue
	select {
	case b.queue <- struct{}{}:
		defer func() { <-b.queue }()
	case <-ctx.Done():
		return ctx.Err()
	default:
		// queue full
		if b.onRejected != nil {
			b.onRejected(ctx)
		}
		return ErrBulkheadRejected
	}

	// Wait for execution slot
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
	case <-ctx.Done():
		return ctx.Err()
	}

	return fn(ctx)
}

func (b *Bulkhead) init() {
	b.once.Do(func() {
		b.sem = make(chan struct{}, b.maxParallel)
		b.queue = make(chan struct{}, b.maxQueue)
	})
}

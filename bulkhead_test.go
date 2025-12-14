package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 并发执行不超过 maxParallel
func TestBulkhead_MaxParallel(t *testing.T) {
	var running atomic.Int32
	maxParallel := 5
	bh := NewBulkhead(maxParallel, 0)

	total := 20
	var wg sync.WaitGroup
	wg.Add(total)

	errs := make([]error, total)
	for i := 0; i < total; i++ {
		go func(i int) {
			defer wg.Done()
			err := bh.Execute(context.Background(), func(ctx context.Context) error {
				curr := running.Add(1)
				if curr > int32(maxParallel) {
					t.Errorf("parallel limit exceeded: %d > %d", curr, maxParallel)
				}
				time.Sleep(10 * time.Millisecond)
				running.Add(-1)
				return nil
			})
			errs[i] = err
		}(i)
	}

	wg.Wait()

	for _, err := range errs {
		if err != nil && err != ErrBulkheadRejected {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

// 队列满时被拒绝 + 回调触发
func TestBulkhead_QueueRejected(t *testing.T) {
	maxParallel := 1
	maxQueue := 2
	var rejected atomic.Int32

	bh := NewBulkhead(maxParallel, maxQueue).OnRejected(func(ctx context.Context) {
		rejected.Add(1)
	})

	total := 10
	var wg sync.WaitGroup
	wg.Add(total)

	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			_ = bh.Execute(context.Background(), func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})
		}()
	}

	wg.Wait()

	if rejected.Load() == 0 {
		t.Fatalf("expected some rejections due to queue full")
	}
}

// 上下文取消
func TestBulkhead_ContextCancel(t *testing.T) {
	bh := NewBulkhead(1, 1)

	start := make(chan struct{})
	done := make(chan struct{})

	// 占用一个 slot
	go func() {
		_ = bh.Execute(context.Background(), func(ctx context.Context) error {
			close(start) // 通知占用 slot
			time.Sleep(50 * time.Millisecond)
			close(done)
			return nil
		})
	}()

	// 等待 slot 占用
	<-start

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	// 这个调用应该因为 context 超时而返回
	err := bh.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context error, got %v", err)
	}

	<-done // 等待占用 goroutine 完成
}

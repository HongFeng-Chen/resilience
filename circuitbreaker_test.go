package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 成功路径，保持 Closed 状态
func TestCircuitBreaker_Success(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)

	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.state != Closed {
		t.Fatalf("expected state Closed, got %s", cb.state)
	}
}

// 连续失败触发 Open
func TestCircuitBreaker_TripOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	failFn := func(ctx context.Context) error { return errors.New("fail") }

	_ = cb.Execute(context.Background(), failFn)
	_ = cb.Execute(context.Background(), failFn)

	if cb.state != Open {
		t.Fatalf("expected state Open, got %s", cb.state)
	}

	// 再执行应返回 ErrCircuitOpen
	err := cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

// BreakDuration 后进入 HalfOpen
func TestCircuitBreaker_HalfOpenTransition(t *testing.T) {
	cb := NewCircuitBreaker(1, 20*time.Millisecond)

	_ = cb.Execute(context.Background(), func(ctx context.Context) error { return errors.New("fail") })
	if cb.state != Open {
		t.Fatalf("expected Open, got %s", cb.state)
	}

	time.Sleep(25 * time.Millisecond)

	err := cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.state != HalfOpen && cb.state != Closed {
		t.Fatalf("expected HalfOpen or Closed, got %s", cb.state)
	}
}

// HalfOpen 成功后 Reset
func TestCircuitBreaker_HalfOpenReset(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)

	_ = cb.Execute(context.Background(), func(ctx context.Context) error { return errors.New("fail") })
	time.Sleep(15 * time.Millisecond)

	err := cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cb.state != Closed {
		t.Fatalf("expected Closed after HalfOpen success, got %s", cb.state)
	}
}

// 回调触发
func TestCircuitBreaker_Callbacks(t *testing.T) {
	breakCalled := int32(0)
	resetCalled := int32(0)
	halfOpenCalled := int32(0)

	cb := NewCircuitBreaker(1, 10*time.Millisecond).
		OnBreak(func(err error, dur time.Duration) { atomic.AddInt32(&breakCalled, 1) }).
		OnReset(func() { atomic.AddInt32(&resetCalled, 1) }).
		OnHalfOpen(func() { atomic.AddInt32(&halfOpenCalled, 1) })

	_ = cb.Execute(context.Background(), func(ctx context.Context) error { return errors.New("fail") })
	if atomic.LoadInt32(&breakCalled) != 1 {
		t.Fatalf("OnBreak not called")
	}

	time.Sleep(15 * time.Millisecond)
	_ = cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
	if atomic.LoadInt32(&halfOpenCalled) != 1 {
		t.Fatalf("OnHalfOpen not called")
	}

	if atomic.LoadInt32(&resetCalled) != 1 {
		t.Fatalf("OnReset not called after HalfOpen success")
	}
}

// 并发安全测试
func TestCircuitBreaker_Concurrent(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	failFn := func(ctx context.Context) error { return errors.New("fail") }

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cb.Execute(context.Background(), failFn)
		}()
	}
	wg.Wait()

	if cb.state != Open {
		t.Fatalf("expected Open after concurrent failures, got %s", cb.state)
	}
}

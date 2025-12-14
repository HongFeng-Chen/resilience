package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// Optimistic：正常完成（无sleep）
func TestTimeout_Optimistic_Success(t *testing.T) {
	to := NewTimeout(50 * time.Millisecond)

	err := to.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Optimistic：函数超时（sleep）
func TestTimeout_Optimistic_Timeout(t *testing.T) {
	to := NewTimeout(10 * time.Millisecond)

	err := to.Execute(context.Background(), func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

// Optimistic：context.DeadlineExceeded（无sleep）
func TestTimeout_Optimistic_DeadlineExceeded(t *testing.T) {
	to := NewTimeout(50 * time.Millisecond)

	err := to.Execute(context.Background(), func(ctx context.Context) error {
		return context.DeadlineExceeded
	})

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout")
	}
}

// Optimistic：context取消（无sleep）
func TestTimeout_Optimistic_ContextCanceled(t *testing.T) {
	to := NewTimeout(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := to.Execute(ctx, func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// Optimistic：OnTimeout 被调用
func TestTimeout_Optimistic_OnTimeoutCalled(t *testing.T) {
	to := NewTimeout(10 * time.Millisecond)

	var called int32
	to.OnTimeout(func(elapsed time.Duration, ctx context.Context) {
		atomic.AddInt32(&called, 1)
	})

	err := to.Execute(context.Background(), func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout")
	}

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected OnTimeout to be called once")
	}
}

// Pessimistic：正常完成
func TestTimeout_Pessimistic_Success(t *testing.T) {
	to := NewTimeout(50 * time.Millisecond).
		WithMode(Pessimistic)

	err := to.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Pessimistic：函数卡死 → 超时
func TestTimeout_Pessimistic_Timeout(t *testing.T) {
	to := NewTimeout(10 * time.Millisecond).
		WithMode(Pessimistic)

	err := to.Execute(context.Background(), func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout")
	}
}

// Pessimistic：外部 Context 取消
func TestTimeout_Pessimistic_ContextCanceled(t *testing.T) {
	to := NewTimeout(100 * time.Millisecond).
		WithMode(Pessimistic)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := to.Execute(ctx, func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled")
	}
}

// Pessimistic：OnTimeout 被调用
func TestTimeout_Pessimistic_OnTimeoutCalled(t *testing.T) {
	to := NewTimeout(10 * time.Millisecond).
		WithMode(Pessimistic)

	var called int32
	to.OnTimeout(func(elapsed time.Duration, ctx context.Context) {
		atomic.AddInt32(&called, 1)
	})

	err := to.Execute(context.Background(), func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout")
	}

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected OnTimeout to be called once")
	}
}

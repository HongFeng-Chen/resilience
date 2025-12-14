package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func failNTimes(n int32) Func {
	var count int32
	return func(ctx context.Context) error {
		if atomic.AddInt32(&count, 1) <= n {
			return errors.New("fail")
		}
		return nil
	}
}

type fakeBackoff struct{}

func (f fakeBackoff) Duration(attempt int) time.Duration {
	return 0
}

// 成功不重试的测试用例
func TestRetry_Success_NoRetry(t *testing.T) {
	r := NewRetry(3)

	calls := int32(0)
	err := r.Execute(context.Background(), func(ctx context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

// 失败后重试成功的测试用例
func TestRetry_RetryThenSuccess(t *testing.T) {
	r := NewRetry(3).
		WithBackoff(fakeBackoff{})

	err := r.Execute(context.Background(), failNTimes(2))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

// 超过最大重试次数
func TestRetry_ExceedMaxRetries(t *testing.T) {
	r := NewRetry(2).
		WithBackoff(fakeBackoff{})

	err := r.Execute(context.Background(), failNTimes(5))
	if err == nil {
		t.Fatalf("expected error")
	}
}

// Forever 一直重试（受控）
func TestRetry_Forever(t *testing.T) {
	r := Forever().
		WithBackoff(fakeBackoff{})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := r.Execute(ctx, func(ctx context.Context) error {
		return errors.New("fail")
	})

	if err == nil {
		t.Fatalf("expected context error")
	}
}

// Handle 条件生效
func TestRetry_HandlePredicate(t *testing.T) {
	r := NewRetry(3).
		Handle(func(err error) bool {
			return err.Error() == "retryable"
		}).
		WithBackoff(fakeBackoff{})

	err := r.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fatal")
	})

	if err == nil || err.Error() != "fatal" {
		t.Fatalf("expected fatal error")
	}
}

// OnRetry 被调用
func TestRetry_OnRetryCalled(t *testing.T) {
	r := NewRetry(3).
		WithBackoff(fakeBackoff{})

	var retries int32
	r.OnRetry(func(attempt int, err error, delay time.Duration, ctx context.Context) {
		atomic.AddInt32(&retries, 1)
	})

	_ = r.Execute(context.Background(), failNTimes(2))

	if retries != 2 {
		t.Fatalf("expected 2 retries, got %d", retries)
	}
}

// Context 取消
func TestRetry_ContextCancelled(t *testing.T) {
	r := Forever().
		WithBackoff(fakeBackoff{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := r.Execute(ctx, func(ctx context.Context) error {
		return errors.New("fail")
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled")
	}
}

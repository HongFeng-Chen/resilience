package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

// 正常返回，不触发 fallback
func TestFallback_NoError(t *testing.T) {
	called := int32(0)
	f := NewFallback(func(ctx context.Context) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	err := f.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("fallback should not be called")
	}
}

// 出错且触发 fallback
func TestFallback_ErrorTriggersFallback(t *testing.T) {
	called := int32(0)
	fallback := func(ctx context.Context) error {
		atomic.AddInt32(&called, 1)
		return nil
	}

	f := NewFallback(fallback)

	err := f.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("fallback should be called once")
	}
}

// 不满足条件不触发 fallback
func TestFallback_HandlePredicate(t *testing.T) {
	called := int32(0)
	fallback := func(ctx context.Context) error {
		atomic.AddInt32(&called, 1)
		return nil
	}

	f := NewFallback(fallback).
		Handle(func(err error) bool {
			return false // 永远不触发
		})

	err := f.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected original error")
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("fallback should not be called")
	}
}

// OnFallback 被调用
func TestFallback_OnFallbackCalled(t *testing.T) {
	var fallbackErr error
	var ctxSeen context.Context

	fallback := func(ctx context.Context) error {
		return nil
	}

	f := NewFallback(fallback).
		OnFallback(func(err error, ctx context.Context) {
			fallbackErr = err
			ctxSeen = ctx
		})

	testErr := errors.New("fail")
	ctx := context.Background()

	_ = f.Execute(ctx, func(ctx context.Context) error {
		return testErr
	})

	if fallbackErr != testErr {
		t.Fatalf("OnFallback should receive original error")
	}
	if ctxSeen != ctx {
		t.Fatalf("OnFallback should receive correct context")
	}
}

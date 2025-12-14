package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockPolicy struct {
	id    string
	trace *[]string
	err   error
}

func (m *mockPolicy) Execute(ctx context.Context, fn Func) error {
	*m.trace = append(*m.trace, m.id+"-before")

	if m.err != nil {
		return m.err
	}

	err := fn(ctx)
	*m.trace = append(*m.trace, m.id+"-after")
	return err
}

// 正确执行顺序（核心测试）
func TestWrapPolicy_ExecutionOrder(t *testing.T) {
	trace := []string{}

	p1 := &mockPolicy{id: "A", trace: &trace}
	p2 := &mockPolicy{id: "B", trace: &trace}
	p3 := &mockPolicy{id: "C", trace: &trace}

	wrap := Wrap(p1, p2, p3)

	err := wrap.Execute(context.Background(), func(ctx context.Context) error {
		trace = append(trace, "fn")
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"A-before",
		"B-before",
		"C-before",
		"fn",
		"C-after",
		"B-after",
		"A-after",
	}

	if len(trace) != len(expected) {
		t.Fatalf("trace length mismatch: %v", trace)
	}

	for i := range expected {
		if trace[i] != expected[i] {
			t.Fatalf("at %d expected %s got %s", i, expected[i], trace[i])
		}
	}
}

// 内层返回错误 → 外层短路
func TestWrapPolicy_ErrorShortCircuit(t *testing.T) {
	trace := []string{}
	sentinel := errors.New("boom")

	p1 := &mockPolicy{id: "A", trace: &trace}
	p2 := &mockPolicy{id: "B", trace: &trace, err: sentinel}
	p3 := &mockPolicy{id: "C", trace: &trace}

	wrap := Wrap(p1, p2, p3)

	err := wrap.Execute(context.Background(), func(ctx context.Context) error {
		trace = append(trace, "fn")
		return nil
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error")
	}

	expected := []string{
		"A-before",
		"B-before",
	}

	for i := range expected {
		if trace[i] != expected[i] {
			t.Fatalf("unexpected trace: %v", trace)
		}
	}
}

// Context 透传正确
func TestWrapPolicy_ContextPropagation(t *testing.T) {
	key := struct{}{}
	value := "hello"

	ctx := context.WithValue(context.Background(), key, value)

	var seen atomic.Int32

	p := &mockPolicy{
		id:    "A",
		trace: &[]string{},
	}

	wrap := Wrap(p)

	err := wrap.Execute(ctx, func(ctx context.Context) error {
		if ctx.Value(key) == value {
			seen.Store(1)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if seen.Load() != 1 {
		t.Fatalf("context value not propagated")
	}
}

// Wrap 空策略 = 直接执行
func TestWrapPolicy_Empty(t *testing.T) {
	called := false

	wrap := Wrap()

	err := wrap.Execute(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if !called {
		t.Fatalf("fn should be called")
	}
}

// WrapPolicy 组合测试
func TestWrapPolicy_Combination(t *testing.T) {
	retry := NewRetry(2)
	timeout := NewTimeout(50 * time.Millisecond) // 函数睡10ms，保证不超时
	fallback := NewFallback(successFunc).Handle(func(err error) bool {
		return true // 处理所有错误
	})

	wrapped := Wrap(retry, timeout, fallback)

	err := wrapped.Execute(context.Background(), func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return errTest
	})

	if err != nil {
		t.Fatalf("expected fallback to handle error, got %v", err)
	}
}

// 并发安全组合测试
func TestWrapPolicy_Concurrent(t *testing.T) {
	retry := NewRetry(1)
	bulkhead := NewBulkhead(3, 3)
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	wrapped := Wrap(retry, bulkhead, cb)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = wrapped.Execute(context.Background(), func(ctx context.Context) error {
				time.Sleep(5 * time.Millisecond)
				return errTest
			})
		}()
	}
	wg.Wait()
}

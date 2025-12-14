package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errTest = errors.New("test error")

func successFunc(ctx context.Context) error {
	return nil
}

func failFunc(ctx context.Context) error {
	return errTest
}

// 空 Wrap（理论最优）
func BenchmarkWrapPolicy_Empty(b *testing.B) {
	wrap := Wrap()

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrap.Execute(context.Background(), fn)
	}
}

// 3 层策略（典型）
func BenchmarkWrapPolicy_3Layers(b *testing.B) {
	p := &mockPolicy{id: "A", trace: &[]string{}}

	wrap := Wrap(p, p, p)

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrap.Execute(context.Background(), fn)
	}
}

// 10 层策略（极限情况）
func BenchmarkWrapPolicy_10Layers(b *testing.B) {
	p := &mockPolicy{id: "A", trace: &[]string{}}

	policies := make([]Resilience, 10)
	for i := range policies {
		policies[i] = p
	}

	wrap := Wrap(policies...)

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrap.Execute(context.Background(), fn)
	}
}

// 串联执行策略（典型）
func BenchmarkWrapPolicy_Serial(b *testing.B) {
	retry := NewRetry(2)
	timeout := NewTimeout(10 * time.Microsecond)
	fallback := NewFallback(successFunc)
	bulkhead := NewBulkhead(5, 5)
	cb := NewCircuitBreaker(3, 50*time.Millisecond)

	wrapped := Wrap(retry, timeout, fallback, bulkhead, cb)

	fn := func(ctx context.Context) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapped.Execute(context.Background(), fn)
	}
}

// 并发执行策略（典型）
func BenchmarkWrapPolicy_Parallel(b *testing.B) {
	retry := NewRetry(2)
	timeout := NewTimeout(10 * time.Microsecond)
	fallback := NewFallback(successFunc)
	bulkhead := NewBulkhead(5, 5)
	cb := NewCircuitBreaker(3, 50*time.Millisecond)

	wrapped := Wrap(retry, timeout, fallback, bulkhead, cb)

	fn := func(ctx context.Context) error { return nil }

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = wrapped.Execute(context.Background(), fn)
		}
	})
}

package resilience

import (
	"context"
	"errors"
	"testing"
)

// 成功路径（无 fallback）
func BenchmarkFallback_Success(b *testing.B) {
	f := NewFallback(func(ctx context.Context) error { return nil })

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Execute(context.Background(), fn)
	}
}

// 错误触发 fallback
func BenchmarkFallback_Fallback(b *testing.B) {
	f := NewFallback(func(ctx context.Context) error { return nil })

	fn := func(ctx context.Context) error {
		return errors.New("fail")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Execute(context.Background(), fn)
	}
}

// Predicate 阻止 fallback 执行
func BenchmarkFallback_PredicateBlocks(b *testing.B) {
	f := NewFallback(func(ctx context.Context) error { return nil }).
		Handle(func(err error) bool { return false })

	fn := func(ctx context.Context) error {
		return errors.New("fail")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Execute(context.Background(), fn)
	}
}

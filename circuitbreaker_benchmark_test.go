package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

// 成功路径
func BenchmarkCircuitBreaker_Success(b *testing.B) {
	cb := NewCircuitBreaker(5, 100*time.Millisecond)

	fn := func(ctx context.Context) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(context.Background(), fn)
	}
}

// 连续失败触发 Open
func BenchmarkCircuitBreaker_Failure(b *testing.B) {
	cb := NewCircuitBreaker(1, 100*time.Millisecond)

	fn := func(ctx context.Context) error { return errors.New("fail") }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(context.Background(), fn)
	}
}

// 并发基准
func BenchmarkCircuitBreaker_Parallel(b *testing.B) {
	cb := NewCircuitBreaker(5, 100*time.Millisecond)

	fn := func(ctx context.Context) error { return nil }

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.Execute(context.Background(), fn)
		}
	})
}

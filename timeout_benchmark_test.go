package resilience

import (
	"context"
	"testing"
	"time"
)

// Optimistic：成功路径（最重要）
func BenchmarkTimeout_Optimistic_Success(b *testing.B) {
	to := NewTimeout(100 * time.Millisecond)

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = to.Execute(context.Background(), fn)
	}
}

// Optimistic：超时路径（受控）
func BenchmarkTimeout_Optimistic_Timeout(b *testing.B) {
	to := NewTimeout(1 * time.Nanosecond)

	fn := func(ctx context.Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = to.Execute(context.Background(), fn)
	}
}

// Pessimistic：成功路径上，悲观模式比乐观模式的延迟更高。
func BenchmarkTimeout_Pessimistic_Success(b *testing.B) {
	to := NewTimeout(100 * time.Millisecond).
		WithMode(Pessimistic)

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = to.Execute(context.Background(), fn)
	}
}

// Pessimistic：超时路径上，悲观模式比乐观模式的延迟更低。

func BenchmarkTimeout_Pessimistic_Timeout(b *testing.B) {
	to := NewTimeout(1 * time.Nanosecond).
		WithMode(Pessimistic)

	fn := func(ctx context.Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = to.Execute(context.Background(), fn)
	}
}

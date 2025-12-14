package resilience

import (
	"context"
	"errors"
	"testing"
)

// 成功路径
func BenchmarkRetry_Success(b *testing.B) {
	r := NewRetry(3)

	fn := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Execute(context.Background(), fn)
	}
}

// 失败但不重试
func BenchmarkRetry_FailNoRetry(b *testing.B) {
	r := NewRetry(3).
		Handle(func(err error) bool {
			return false
		})

	fn := func(ctx context.Context) error {
		return errors.New("fail")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Execute(context.Background(), fn)
	}
}

// 多次重试
func BenchmarkRetry_Retry3Times(b *testing.B) {
	r := NewRetry(3).
		WithBackoff(fakeBackoff{})

	fn := failNTimes(3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Execute(context.Background(), fn)
	}
}

// Forever（受控）
func BenchmarkRetry_Forever(b *testing.B) {
	r := Forever().
		WithBackoff(fakeBackoff{})

	ctx, cancel := context.WithCancel(context.Background())

	fn := func(ctx context.Context) error {
		cancel()
		return errors.New("fail")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Execute(ctx, fn)
	}
}

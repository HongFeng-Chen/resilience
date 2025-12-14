package resilience

import (
	"context"
	"testing"
	"time"
)

// 单线程基准
func BenchmarkBulkhead_Serial(b *testing.B) {
	bh := NewBulkhead(10, 10)

	fn := func(ctx context.Context) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bh.Execute(context.Background(), fn)
	}
}

// 并发基准（模拟高 QPS）
func BenchmarkBulkhead_Parallel(b *testing.B) {
	bh := NewBulkhead(50, 100)

	fn := func(ctx context.Context) error {
		time.Sleep(time.Microsecond) // 模拟轻量任务
		return nil
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bh.Execute(context.Background(), fn)
		}
	})
}

// 队列满 + 拒绝基准
func BenchmarkBulkhead_QueueRejected(b *testing.B) {
	bh := NewBulkhead(1, 1)

	fn := func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bh.Execute(context.Background(), fn)
		}
	})
}

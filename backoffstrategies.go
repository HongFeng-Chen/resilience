package resilience

import (
	"math/rand"
	"time"
)

type BackoffStrategy interface {
	Duration(attempt int) time.Duration
}

// NoBackoff
// 无延迟策略，每次重试间隔为0秒。
type NoBackoff struct{}

func (NoBackoff) Duration(_ int) time.Duration {
	return 0
}

// FixedBackoff
// 固定延迟策略，每次重试间隔为固定的时间。
type FixedBackoff struct {
	Delay time.Duration
}

func (b FixedBackoff) Duration(_ int) time.Duration {
	return b.Delay
}

// ExponentialBackoff
// 指数延迟策略，每次重试间隔为前一次的两倍。
type ExponentialBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// Duration returns the exponential backoff duration for the given attempt.
func (b ExponentialBackoff) Duration(attempt int) time.Duration {
	d := b.BaseDelay * (1 << (attempt - 1))
	if d > b.MaxDelay {
		return b.MaxDelay
	}
	return d
}

// JitterBackoff (Exponential + Full Jitter)
// 抖动延迟策略，每次重试间隔为指数延迟策略加上随机延迟。
type JitterBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// Duration returns a random duration between 0 and the exponential backoff delay.
func (b JitterBackoff) Duration(attempt int) time.Duration {
	max := b.BaseDelay * (1 << (attempt - 1))
	if max > b.MaxDelay {
		max = b.MaxDelay
	}
	return time.Duration(rand.Int63n(int64(max)))
}

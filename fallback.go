package resilience

import (
	"context"
)

type Fallback struct {
	shouldFallback func(error) bool
	fallbackFunc   Func
	onFallback     OnFallbackFunc
}

type OnFallbackFunc func(err error, ctx context.Context)

// NewFallback creates a fallback policy
func NewFallback(fallback Func) *Fallback {
	return &Fallback{
		fallbackFunc: fallback,
		shouldFallback: func(err error) bool {
			return err != nil
		},
	}
}

func (f *Fallback) Handle(fn func(error) bool) *Fallback {
	f.shouldFallback = fn
	return f
}

func (f *Fallback) OnFallback(fn OnFallbackFunc) *Fallback {
	f.onFallback = fn
	return f
}

func (f *Fallback) Execute(ctx context.Context, fn Func) error {
	err := fn(ctx)
	if err == nil {
		return nil
	}

	if !f.shouldFallback(err) {
		return err
	}

	if f.onFallback != nil {
		f.onFallback(err, ctx)
	}

	return f.fallbackFunc(ctx)
}

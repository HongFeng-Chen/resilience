package resilience

import "context"

type Resilience interface {
	Execute(ctx context.Context, fn Func) error
}

type Func func(ctx context.Context) error

package resilience

import (
	"context"
)

// WrapPolicy allows multiple resilience policies to be composed
// Execution order:
//
//	Wrap(A, B, C).Execute(fn)
//	=> A.Execute(B.Execute(C.Execute(fn)))
type WrapPolicy struct {
	policies []Resilience
}

// Wrap composes multiple policies into a single policy
func Wrap(policies ...Resilience) *WrapPolicy {
	return &WrapPolicy{
		policies: policies,
	}
}

func (w *WrapPolicy) Execute(ctx context.Context, fn Func) error {
	if len(w.policies) == 0 {
		return fn(ctx)
	}

	// Build execution chain from inside out
	wrapped := fn

	for i := len(w.policies) - 1; i >= 0; i-- {
		policy := w.policies[i]
		next := wrapped

		wrapped = func(ctx context.Context) error {
			return policy.Execute(ctx, next)
		}
	}

	return wrapped(ctx)
}

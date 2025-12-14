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
	chain    Func // prebuilt execution chain, except for the innermost fn
}

// Wrap composes multiple policies into a single policy
// Wrap composes multiple policies into a single policy
func Wrap(policies ...Resilience) *WrapPolicy {
	wp := &WrapPolicy{
		policies: policies,
	}

	if len(policies) == 0 {
		// No policies, chain is nil
		wp.chain = nil
		return wp
	}

	// Build a pre-wrapped chain using a dummy innermost function
	dummyFn := func(ctx context.Context) error {
		return nil
	}

	wrapped := dummyFn
	for i := len(policies) - 1; i >= 0; i-- {
		policy := policies[i]
		next := wrapped
		wrapped = func(ctx context.Context) error {
			return policy.Execute(ctx, next)
		}
	}

	wp.chain = wrapped
	return wp
}

// Execute runs the wrapped chain with the actual innermost function fn
func (w *WrapPolicy) Execute(ctx context.Context, fn Func) error {
	if w.chain == nil {
		// No policies, execute fn directly
		return fn(ctx)
	}

	// Replace the innermost dummy function with the actual fn
	exec := fn
	for i := len(w.policies) - 1; i >= 0; i-- {
		policy := w.policies[i]
		next := exec
		exec = func(ctx context.Context) error {
			return policy.Execute(ctx, next)
		}
	}

	return exec(ctx)
}

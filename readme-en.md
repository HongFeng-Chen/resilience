# retry

A **Polly-inspired Resilience Library for Go**.

This package provides a production-ready implementation of common resilience patterns, closely mirroring the design and behavior of **C# Polly**, but written in idiomatic Go and built around `context.Context`.

---

## ✨ Features

* ✅ Retry (fixed / exponential / jitter / forever)
* ✅ Circuit Breaker (Open / Half-Open / Closed)
* ✅ Timeout (Optimistic & Pessimistic)
* ✅ Fallback
* ✅ Bulkhead (concurrency & queue isolation)
* ✅ Policy Wrap (strategy composition)
* ✅ Context-aware, goroutine-safe
* ✅ Production ready

---

## 📦 Installation

```bash
go get  github.com/HongFeng-Chen/resilience
```

---

## 🧠 Core Concepts

All policies implement the same interface:

```go
type Resilience interface {
    Execute(ctx context.Context, fn Func) error
}

type Func func(ctx context.Context) error
```

This allows **any policy to be freely composed** with others.

---

## 🔁 Retry

Retry a function when it fails.

```go
policy := resilience.NewRetry(3).
    Handle(func(err error) bool {
        return errors.Is(err, ErrTransient)
    }).
    WithBackoff(resilience.JitterBackoff{
        BaseDelay: time.Second,
        MaxDelay:  5 * time.Second,
    }).
    OnRetry(func(attempt int, err error, delay time.Duration, ctx context.Context) {
        log.Printf("retry #%d after %v", attempt, delay)
    })

err := policy.Execute(ctx, callAPI)
```

### Retry Forever

```go
policy := resilience.Forever()
```

---

## 🔌 Circuit Breaker

Stops calling a failing dependency.

```go
breaker := resilience.NewCircuitBreaker(3, 10*time.Second).
    OnBreak(func(err error, d time.Duration) {
        log.Println("breaker opened")
    }).
    OnHalfOpen(func() {
        log.Println("breaker half-open")
    }).
    OnReset(func() {
        log.Println("breaker reset")
    })

err := breaker.Execute(ctx, callAPI)
```

---

## ⏱ Timeout

Fail fast when execution takes too long.

### Optimistic (recommended)

```go
timeout := resilience.NewTimeout(2 * time.Second)

err := timeout.Execute(ctx, callAPI)
```

### Pessimistic

```go
timeout := resilience.NewTimeout(2 * time.Second).
    WithMode(retry.Pessimistic)
```

---

## 🧯 Fallback

Execute an alternative action when failures occur.

```go
fallback := resilience.NewFallback(func(ctx context.Context) error {
    return useCache()
}).OnFallback(func(err error, ctx context.Context) {
    log.Println("fallback triggered")
})

err := fallback.Execute(ctx, callAPI)
```

---

## 🚢 Bulkhead

Limit concurrency and protect downstream resources.

```go
bulkhead := resilience.NewBulkhead(10, 50).
    OnRejected(func(ctx context.Context) {
        log.Println("bulkhead rejected")
    })

err := bulkhead.Execute(ctx, callAPI)
```

---

## 🧩 Wrap (Policy Composition)

Combine multiple policies into one.

```go
policy := resilience.Wrap(
    resilience.NewBulkhead(20, 100),
    resilience.New(3),
    resilience.NewCircuitBreaker(5, 30*time.Second),
    resilience.NewTimeout(2*time.Second),
    resilience.NewFallback(func(ctx context.Context) error {
        return useCache()
    }),
)

err := policy.Execute(ctx, callAPI)
```

### Execution Order

```text
Bulkhead
 └── Retry
     └── CircuitBreaker
         └── Timeout
             └── Fallback
                 └── callAPI
```

> The **first policy is always the outermost**.

---

## 🧪 Error Handling

Common exported errors:

```go
resilience.ErrTimeout
resilience.ErrCircuitOpen
resilience.ErrBulkheadRejected
```

---

## 🏗 Design Principles

* Inspired by **C# Polly**, not a direct port
* Idiomatic Go
* `context.Context` first
* No hidden goroutines (except pessimistic timeout)
* Explicit composition over magic

---

## 📈 Recommended Production Setup

```go
policy := resilience.Wrap(
    resilience.NewBulkhead(50, 200),
    resilience.New(3).
        WithBackoff(resilience.JitterBackoff{
            BaseDelay: 500 * time.Millisecond,
            MaxDelay:  5 * time.Second,
        }),
    resilience.NewCircuitBreaker(5, 30*time.Second),
    resilience.NewTimeout(2*time.Second),
    resilience.NewFallback(func(ctx context.Context) error {
        return useCache()
    }),
)
```

---

## 🧭 Roadmap

* [ ] Metrics / OpenTelemetry hooks
* [ ] Async helpers
* [ ] Generic Result Policies
* [ ] Examples & benchmarks

---

## 📜 License

MIT License

---

## 🙌 Acknowledgements

Inspired by:

* [Polly](https://github.com/App-vNext/Polly)
* Go resilience best practices

# resilience

一个 **参考 C# Polly 的 Go 弹性库**。

此库提供了生产级的常用弹性策略实现,使用 Go idiomatic 风格，并以 `context.Context` 为中心。

---

## ✨ 功能特性

* ✅ Retry（固定 / 指数 / Jitter / 永远重试）
* ✅ Circuit Breaker（打开 / 半开 / 关闭）
* ✅ Timeout（乐观 & 悲观）
* ✅ Fallback（降级处理）
* ✅ Bulkhead（并发与队列隔离）
* ✅ Policy Wrap（策略组合）
* ✅ 支持 Context，goroutine 安全
* ✅ 可用于生产环境

---

## 📦 安装

```bash
go get github.com/your-org/resilience
```

---

## 🧠 核心概念

所有策略都实现相同接口：

```go
type Resilience interface {
    Execute(ctx context.Context, fn Func) error
}

type Func func(ctx context.Context) error
```

这样可以**任意策略自由组合**。

---

## 🔁 Retry（重试）

在函数执行失败时进行重试。

```go
func main() {
	
	subResult := 0
	err := resilience.NewRetry(3).
		Handle(func(err error) bool {
			return errors.Is(err, ErrMyCustom) 
		}).
		WithBackoff(resilience.FixedBackoff{
			Delay: 2 * time.Second, 
		}).
		OnRetry(func(attempt int, err error, delay time.Duration, ctx context.Context) {
			log.Printf("%s: 第 %d 次重试, 延迟 %v", time.Now().Format("2006-01-02 15:04:05.000"), attempt, delay)
		}).Execute(context.Background(), func(ctx context.Context) error {
		var suberr error
		subResult, suberr = doSomething2(ctx)
		return suberr
	})
	fmt.Println("subResult:", subResult)

	if err != nil {
		log.Println("最终失败:", err)
	}

}

var ErrMyCustom = errors.New("my custom error")

func doSomething(ctx context.Context) error {
	return fmt.Errorf("wrapped: %w", ErrMyCustom)
}
func doSomething2(ctx context.Context) (int, error) {
	return 42, nil
}
```

### 永久重试

```go
policy := resilience.Forever()
```

---

## 🔌 Circuit Breaker（熔断器）

在依赖失败时停止调用。

```go
breaker := resilience.NewCircuitBreaker(3, 10*time.Second).
    OnBreak(func(err error, d time.Duration) {
        log.Println("熔断器打开")
    }).
    OnHalfOpen(func() {
        log.Println("熔断器半开")
    }).
    OnReset(func() {
        log.Println("熔断器重置")
    })

err := breaker.Execute(ctx, callAPI)
```

---

## ⏱ Timeout（超时）

在执行超时时快速失败。

### 乐观模式（推荐）

```go
timeout := resilience.NewTimeout(2 * time.Second)

err := timeout.Execute(ctx, callAPI)
```

### 悲观模式

```go
timeout := resilience.NewTimeout(2 * time.Second).
    WithMode(resilience.Pessimistic)
```

---

## 🧯 Fallback（降级）

当执行失败时执行备用逻辑。

```go
fallback := resilience.NewFallback(func(ctx context.Context) error {
    return useCache()
}).OnFallback(func(err error, ctx context.Context) {
    log.Println("触发降级逻辑")
})

err := fallback.Execute(ctx, callAPI)
```

---

## 🧩 Bulkhead（舱壁隔离）

限制并发数量，保护下游资源。

```go
bulkhead := resilience.NewBulkhead(10, 50).
    OnRejected(func(ctx context.Context) {
        log.Println("bulkhead 拒绝执行")
    })

err := bulkhead.Execute(ctx, callAPI)
```

---

## 🧩 Wrap（策略组合）

将多个策略组合成一个。

```go
policy := resilience.Wrap(
    resilience.NewBulkhead(20, 100),
    resilience.NewRetry(3),
    resilience.NewCircuitBreaker(5, 30*time.Second),
    resilience.NewTimeout(2*time.Second),
    resilience.NewFallback(func(ctx context.Context) error {
        return useCache()
    }),
)

err := policy.Execute(ctx, callAPI)
```

### 执行顺序

```text
Bulkhead
 └── Retry
     └── CircuitBreaker
         └── Timeout
             └── Fallback
                 └── callAPI
```

> 第一个策略总是最外层。

---

## 🧪 错误处理

常用导出错误：

```go
resilience.ErrTimeout
resilience.ErrCircuitOpen
resilience.ErrBulkheadRejected
```

---

## 🏗 设计原则

* 参考 **Polly**，非直接移植
* 使用 Go idiomatic 风格
* 以 `context.Context` 为中心
* 除悲观超时外无隐藏 goroutine
* 明确组合优于魔法调用

---

## 📈 推荐生产组合

```go
policy := retry.Wrap(
    resilience.NewBulkhead(50, 200),
    resilience.NewRetry(3).
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

## 🧭 发展计划

* [ ] 集成指标 / OpenTelemetry
* [ ] 异步调用支持
* [ ] 泛型结果策略
* [ ] 示例 & 性能基准

---

## 📜 许可证

MIT License

---

## 🙌 致谢

灵感来自：

* [Polly](https://github.com/App-vNext/Polly)
* Go 弹性设计实践

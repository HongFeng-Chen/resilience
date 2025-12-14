
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"resilience"
	"time"
)

func main() {

	subResult := 0
	err := resilience.NewRetry(3).
		Handle(func(err error) bool {
			return errors.Is(err, ErrMyCustom) // 现在是同一个变量！
		}).
		WithBackoff(resilience.FixedBackoff{
			Delay: 2 * time.Second, // 2 seconds
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

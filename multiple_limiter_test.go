package rate

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewMultipleLimiter(t *testing.T) {
	l, err := NewMultipleLimiter(
		NewLimiter(10*time.Second, 10),
		NewLimiter(20*time.Second, 15),
		NewLimiter(30*time.Second, 20),
	)
	if err != nil {
		panic(err)
	}

	step := 0
	for {
		ctx, _ := context.WithTimeout(context.Background(), 61*time.Second)
		err = l.Wait(ctx)
		if err != nil {
			panic(err)
		}
		next := l.Next()

		fmt.Printf("time: %v, step: %v, next: %v\n", time.Now(), step, next)
		step++
	}
}

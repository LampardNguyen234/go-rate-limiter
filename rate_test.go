package rate

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	l := NewLimiter(10*time.Second, 20)

	ctx := context.Background()
	count := 0
	for {
		tmpCtx, _ := context.WithTimeout(ctx, 15*time.Second)
		err := l.Wait(tmpCtx)
		if err != nil {
			panic(err)
		}

		next := l.Next()
		fmt.Printf("time: %v, count: %v, next: %v\n", time.Now(), count, next)
		count++
	}
}

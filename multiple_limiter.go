package rate

import (
	"context"
	"fmt"
	"time"
)

type MultipleLimiter struct {
	limiters []RateLimiter
}

func NewMultipleLimiter(limiters ...RateLimiter) (RateLimiter, error) {
	if len(limiters) == 0 {
		return nil, fmt.Errorf("empty limiter")
	}

	return &MultipleLimiter{
		limiters: append([]RateLimiter{}, limiters...),
	}, nil
}

func (l *MultipleLimiter) Next() time.Time {
	next := l.limiters[0].Next()
	for i, limiter := range l.limiters {
		if i == 0 {
			continue
		}

		if limiter.Next().After(next) {
			next = limiter.Next()
		}
	}

	return next
}

func (l *MultipleLimiter) Allow() bool {
	for _, limiter := range l.limiters {
		if !limiter.Allow() {
			return false
		}
	}
	return true
}

func (l *MultipleLimiter) AllowN(t time.Time, n int) bool {
	for _, limiter := range l.limiters {
		if !limiter.AllowN(t, n) {
			return false
		}
	}
	return true
}

func (l *MultipleLimiter) Wait(ctx context.Context) error {
	errChan := make(chan error)
	retChan := make(chan bool)

	var wait = func(limiter RateLimiter) {
		err := limiter.Wait(ctx)
		if err != nil {
			errChan <- err
		} else {
			retChan <- true
		}
	}

	for _, limiter := range l.limiters {
		go wait(limiter)
	}

	numSuccess := 0
	for {
		select {
		case err := <-errChan:
			return err
		case <-retChan:
			numSuccess += 1
		default:
			if numSuccess == len(l.limiters) {
				return nil
			}
		}
	}
}
func (l *MultipleLimiter) WaitN(ctx context.Context, n int) error {
	errChan := make(chan error)
	retChan := make(chan bool)

	var wait = func(limiter RateLimiter) {
		err := limiter.WaitN(ctx, n)
		if err != nil {
			errChan <- err
		} else {
			retChan <- true
		}
	}

	for _, limiter := range l.limiters {
		go wait(limiter)
	}

	numSuccess := 0
	for {
		select {
		case err := <-errChan:
			return err
		case <-retChan:
			numSuccess += 1
		default:
			if numSuccess == len(l.limiters) {
				return nil
			}
		}
	}
}

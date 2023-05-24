// Package rate provides a rate limiter.
package rate

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// A Limiter controls how frequently events are allowed to happen.
// It implements a "token bucket" of size b, initially full and refilled
// at rate r tokens per second.
// Informally, in any large enough time interval, the Limiter limits the
// rate to r tokens per second, with a maximum burst size of b events.
// As a special case, if r == Inf (the infinite rate), b is ignored.
// See https://en.wikipedia.org/wiki/Token_bucket for more about token buckets.
//
// The zero value is a valid Limiter, but it will reject all events.
// Use NewLimiter to create non-zero Limiters.
//
// Limiter has three main methods, Allow, Reserve, and Wait.
// Most callers should use Wait.
//
// Each of the three methods consumes a single token.
// They differ in their behavior when no token is available.
// If no token is available, Allow returns false.
// If no token is available, Reserve returns a reservation for a future token
// and the amount of time the caller must wait before using it.
// If no token is available, Wait blocks until one can be obtained
// or its associated context.Context is canceled.
//
// The methods AllowN, ReserveN, and WaitN consume n tokens.
type Limiter struct {
	mu       *sync.Mutex
	interval time.Duration

	tokens []time.Time
}

// NewLimiter returns a new Limiter that allows up to b events in an interval.
func NewLimiter(interval time.Duration, b int) *Limiter {
	t := time.Now()
	tokens := make([]time.Time, 0)
	for i := 0; i < b; i++ {
		tokens = append(tokens, t)
	}
	return &Limiter{
		tokens:   tokens,
		interval: interval,
		mu:       new(sync.Mutex),
	}
}

func (lim *Limiter) Next() time.Time {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	return lim.tokens[0]
}

// Allow reports whether an event may happen now.
func (lim *Limiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// AllowN reports whether n events may happen at time t.
// Use this method if you intend to drop / skip events that exceed the rate limit.
// Otherwise use Reserve or Wait.
func (lim *Limiter) AllowN(t time.Time, n int) bool {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	if n > len(lim.tokens) {
		return false
	}

	return t.After(lim.tokens[n-1])
}

// Wait is shorthand for WaitN(ctx, 1).
func (lim *Limiter) Wait(ctx context.Context) (err error) {
	return lim.WaitN(ctx, 1)
}

// WaitN blocks until lim permits n events to happen.
// It returns an error if n exceeds the Limiter's burst size, the Context is
// canceled, or the expected wait time exceeds the Context's Deadline.
// The burst limit is ignored if the rate limit is Inf.
func (lim *Limiter) WaitN(ctx context.Context, n int) error {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	if n == 0 {
		return fmt.Errorf("rate: Wait(n=0) not allowed")
	}

	burst := len(lim.tokens)
	if n > burst {
		return fmt.Errorf("rate: Wait(n=%d) exceeds limiter's burst %d", n, burst)
	}
	target := lim.tokens[n-1]

	deadline, ok := ctx.Deadline()
	if ok {
		if deadline.Before(target) {
			return fmt.Errorf("rate: Wait(n=%d) would exceed context deadline", n)
		}
	}

	delay := target.Sub(time.Now())
	if delay > 0 {
		// The test code calls lim.wait with a fake timer generator.
		// This is the real timer generator.
		newTimer := func(d time.Duration) (<-chan time.Time, func() bool, func()) {
			timer := time.NewTimer(d)
			return timer.C, timer.Stop, func() {}
		}

		ch, stop, advance := newTimer(delay)
		defer stop()
		advance() // only has an effect when testing
		select {
		case <-ch:
			break
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	lim.tokens = lim.tokens[n:]
	for i := 0; i < n; i++ {
		lim.tokens = append(lim.tokens, time.Now().Add(lim.interval))
	}

	return nil
}

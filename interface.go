package rate

import (
	"context"
	"time"
)

// RateLimiter implements a simple RateLimiter.
type RateLimiter interface {
	// Next returns the next timestamp an event can occur.
	Next() time.Time

	// Allow reports whether an event may happen now.
	Allow() bool

	// AllowN reports whether n events may happen at time t.
	// Use this method if you intend to drop / skip events that exceed the rate limit.
	// Otherwise use Reserve or Wait.
	AllowN(time.Time, int) bool

	// Wait is shorthand for WaitN(context.Context, 1).
	Wait(context.Context) error

	// WaitN blocks until lim permits n events to happen.
	// It returns an error if n exceeds the Limiter's burst size, the Context is
	// canceled, or the expected wait time exceeds the Context's Deadline.
	// The burst limit is ignored if the rate limit is Inf.
	WaitN(context.Context, int) error
}

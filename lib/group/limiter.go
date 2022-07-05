package group

import (
	"context"
)

// Limiter represents a type that can limit the number of concurrent operations.
type Limiter interface {
	// Acquire blocks until either the context is canceled, or there is
	// availability to start a new operation. If Acquire does not return an
	// error, the returned `release` must be called when the operation
	// completes. The returned context should be used for the operation so that
	// the limiter can cancel the context if needed. Acquire must not return an
	// error unless the context is canceled.
	Acquire(context.Context) (ctx context.Context, release context.CancelFunc, err error)
}

func NewBasicLimiter(limit int) Limiter {
	return make(basicLimiter, limit)
}

// basicLimiter is a simple limiter that limits the number of concurrent
// operations to a constant limit (the len of the channel)
type basicLimiter chan struct{}

func (l basicLimiter) Acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	select {
	case l <- struct{}{}: // will block if the channel is full
		return ctx, func() { <-l }, nil
	case <-ctx.Done(): // return with an error if the context has been canceled
		return ctx, func() {}, ctx.Err()
	}
}

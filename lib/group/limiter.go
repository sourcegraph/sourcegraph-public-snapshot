package group

import (
	"context"
)

type Limiter interface {
	Acquire(context.Context) (context.Context, context.CancelFunc, error)
}

type unlimitedLimiter struct{}

func (l *unlimitedLimiter) Acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	return ctx, func() {}, nil
}

func newBasicLimiter(limit int) Limiter {
	return make(basicLimiter, limit)
}

type basicLimiter chan struct{}

func (l basicLimiter) Acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	select {
	case l <- struct{}{}:
		return ctx, func() { <-l }, nil
	case <-ctx.Done():
		return ctx, func() {}, ctx.Err()
	}
}

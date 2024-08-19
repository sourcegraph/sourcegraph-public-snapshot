package glock

import (
	"context"
	"sync"
	"time"
)

type ctxKey struct{}

var glockCtxKey = &ctxKey{}

// WithContext returns a context derived from the provided context with
// the provided Clock as a value.
func WithContext(ctx context.Context, clock Clock) context.Context {
	return context.WithValue(ctx, glockCtxKey, clock)
}

// FromContext retrieves the Clock value from the provided context. If a Clock
// is not set on the context a new real clock will be returned.
func FromContext(ctx context.Context) Clock {
	clock, ok := ctx.Value(glockCtxKey).(Clock)
	if !ok {
		clock = NewRealClock()
	}

	return clock
}

type glockAwareContext struct {
	context.Context
	mu   sync.Mutex
	err  error
	done chan struct{}
}

// ContextWithDeadline mimmics context.WithDeadline, but uses the given clock instance
// instead of the using standard time.After function directly.
func ContextWithDeadline(ctx context.Context, clock Clock, deadline time.Time) (context.Context, context.CancelFunc) {
	return ContextWithTimeout(ctx, clock, deadline.Sub(clock.Now()))
}

// ContextWithTimeout mimmics context.WithTimeout, but uses the given clock instance
// instead of the using standard time.After function directly.
func ContextWithTimeout(ctx context.Context, clock Clock, timeout time.Duration) (context.Context, context.CancelFunc) {
	done := make(chan struct{})
	canceled := make(chan struct{})

	ctx, cancel := context.WithCancel(ctx)
	child := &glockAwareContext{Context: ctx, done: done}

	go func() {
		defer cancel()
		defer close(done)

		child.setErr(watchContext(ctx, canceled, clock.After(timeout)))
	}()

	return child, closeOnce(canceled)
}

func (ctx *glockAwareContext) Done() <-chan struct{} {
	return ctx.done
}

func (ctx *glockAwareContext) Err() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.err
}

func (ctx *glockAwareContext) setErr(err error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.err = err
}

func watchContext(ctx context.Context, canceled <-chan struct{}, afterCh <-chan time.Time) error {
	select {
	case <-canceled:
		return context.Canceled
	case <-afterCh:
		return context.DeadlineExceeded
	case <-ctx.Done():
		return ctx.Err()
	}
}

func closeOnce(ch chan<- struct{}) context.CancelFunc {
	var once sync.Once

	return func() {
		once.Do(func() {
			close(ch)
		})
	}
}

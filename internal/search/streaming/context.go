package streaming

import (
	"context"
	"sync"

	"go.uber.org/atomic"
)

type mutValueCtxKey int64

const (
	CanceledLimitHit mutValueCtxKey = iota + 1
)

// WithMutableValue returns a context with a mutable key-value pair. Use the Set
// method to set key and value.
func WithMutableValue(ctx context.Context) *mutValueCtx {
	return &mutValueCtx{Context: ctx}
}

// mutValueCtx is modelled after valueCtx in the standard library, but permits
// to set a value after the context has been created.
type mutValueCtx struct {
	context.Context

	// Mutable key and value.
	key   atomic.Int64
	value atomic.Bool
}

// Set atomically updates key and value stored in ctx.
func (ctx *mutValueCtx) Set(key mutValueCtxKey, value bool) {
	ctx.key.Store(int64(key))
	ctx.value.Store(value)
}

func (ctx *mutValueCtx) Value(key interface{}) interface{} {
	if mutValueCtxKey(ctx.key.Load()) == key {
		return ctx.value.Load()
	}
	return ctx.Context.Value(key)
}

// PickyContext wraps a context and ignores context cancellations if any of the
// parent contexts store a key:value pair with key=reason and value=true. Always
// call defer cleanup() to clean up the goroutine created in Done().
func PickyContext(parent context.Context, reason mutValueCtxKey) (context.Context, func()) {
	done := make(chan struct{})

	// once protects c from being closed twice.
	once := sync.Once{}
	c := make(chan struct{})

	ctx := &pickyCtx{Context: parent, d: done}

	go func() {
		select {
		case <-parent.Done():
			// Check if any parent context has a key:value pair ctx.reason=true.
			val := parent.Value(reason)
			if b, ok := val.(bool); ok && b {
				// "done" will only be closed if the func() returned by PickyContext is called
				// explicitly.
				return
			}
			ctx.mu.Lock()
			ctx.err = parent.Err()
			ctx.mu.Unlock()
		case <-c:
			ctx.mu.Lock()
			if ctx.err == nil {
				ctx.err = context.Canceled
			}
			ctx.mu.Unlock()
		}
		close(done)
	}()
	return ctx, func() { once.Do(func() { close(c) }) }
}

type pickyCtx struct {
	context.Context
	d chan struct{}

	mu  sync.Mutex
	err error
}

func (ctx *pickyCtx) Done() <-chan struct{} {
	return ctx.d
}

func (ctx *pickyCtx) Err() error {
	ctx.mu.Lock()
	err := ctx.err
	ctx.mu.Unlock()
	return err
}

package diskcache

import (
	"context"
	"time"
)

type isolatedTimeoutContext struct {
	parent      context.Context
	deadlineCtx context.Context
}

// withIsolatedTimeout creates a context with a timeout isolated from any timeouts in any of the ancestor contexts.
// Context values are pulled from the parent context only.
func withIsolatedTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	deadlineCtx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	return &isolatedTimeoutContext{
		parent:      parent,
		deadlineCtx: deadlineCtx,
	}, cancelFunc
}

var _ context.Context = &isolatedTimeoutContext{}

func (c *isolatedTimeoutContext) Deadline() (time.Time, bool) {
	return c.deadlineCtx.Deadline()
}

func (c *isolatedTimeoutContext) Done() <-chan struct{} {
	return c.deadlineCtx.Done()
}

func (c *isolatedTimeoutContext) Err() error {
	return c.deadlineCtx.Err()
}

func (c *isolatedTimeoutContext) Value(key any) any {
	return c.parent.Value(key)
}

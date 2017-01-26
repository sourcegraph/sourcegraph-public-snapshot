package langserver

import (
	"context"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
)

// cancel manages $/cancelRequest by keeping track of running commands
type cancel struct {
	mu sync.Mutex
	m  map[jsonrpc2.ID]func()
}

// WithCancel is like context.WithCancel, except you can also cancel via
// calling c.Cancel with the same id.
func (c *cancel) WithCancel(ctx context.Context, id jsonrpc2.ID) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	if c.m == nil {
		c.m = make(map[jsonrpc2.ID]func())
	}
	c.m[id] = cancel
	c.mu.Unlock()
	return ctx, func() {
		c.mu.Lock()
		delete(c.m, id)
		c.mu.Unlock()
		cancel()
	}
}

// Cancel will cancel the request with id. If the request has already been
// cancelled or not been tracked before, Cancel is a noop.
func (c *cancel) Cancel(id jsonrpc2.ID) {
	var cancel func()
	c.mu.Lock()
	if c.m != nil {
		cancel = c.m[id]
		delete(c.m, id)
	}
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

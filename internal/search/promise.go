package search

import (
	"context"
	"sync"
)

type Promise struct {
	getOnce sync.Once
	err     error

	initOnce sync.Once
	done     chan struct{}

	valueOnce sync.Once
	value     interface{}
}

func (p *Promise) init() {
	p.initOnce.Do(func() { p.done = make(chan struct{}) })
}

// Resolve returns a promise that is resolved with a given value.
func (p *Promise) Resolve(v interface{}) *Promise {
	p.valueOnce.Do(func() {
		p.init()
		p.value = v
		close(p.done)
	})
	return p
}

// Get returns the value. It blocks until the promise resolves or the context is
// canceled. Further calls to Get will always return the original results, IE err
// will stay nil even if the context expired between the first and the second
// call. Vice versa, if ctx finishes while resolving, then we will always return
// ctx.Err()
func (p *Promise) Get(ctx context.Context) (interface{}, error) {
	p.getOnce.Do(func() {
		p.init()
		select {
		case <-ctx.Done():
			p.err = ctx.Err()
		case <-p.done:
		}
	})
	return p.value, p.err
}

package search

import (
	"context"
	"sync"
)

type Promise struct {
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
// canceled.
func (p *Promise) Get(ctx context.Context) (interface{}, error) {
	p.init()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.done:
	}
	return p.value, nil
}

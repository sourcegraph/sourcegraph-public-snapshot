package search

import (
	"context"
	"sync"
)

type RepoPromise struct {
	initOnce sync.Once
	done     chan struct{}

	valueOnce sync.Once
	value     []*RepositoryRevisions
}

func (p *RepoPromise) init() {
	p.initOnce.Do(func() { p.done = make(chan struct{}) })
}

// Resolve returns a promise that is resolved with a given value.
func (p *RepoPromise) Resolve(v []*RepositoryRevisions) *RepoPromise {
	p.valueOnce.Do(func() {
		p.init()
		p.value = v
		close(p.done)
	})
	return p
}

// Get returns the resolved revisions. It blocks until the promise resolves or
// the context is canceled.
func (p *RepoPromise) Get(ctx context.Context) ([]*RepositoryRevisions, error) {
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

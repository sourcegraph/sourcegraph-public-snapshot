// Package backend implements the built-in search providers such as indexed
// search and our JIT searcher.
package backend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
)

// Mock is a mock search.Searcher
type Mock struct {
	Result *search.Result
	Error  error

	LastQ    query.Q
	LastOpts *search.Options
}

func (m *Mock) Search(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error) {
	m.LastQ = q
	m.LastOpts = opts
	return m.Result, m.Error
}

func (m *Mock) Close() {}

func (m *Mock) String() string { return "mock" }

type shard struct {
	search.Searcher
	query.Q
	*search.Options
}

func shardedSearch(ctx context.Context, shards <-chan shard) (*search.Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	resC := make(chan struct {
		*search.Result
		error
	})
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := shard.Searcher.Search(ctx, shard.Q, shard.Options)
			resC <- struct {
				*search.Result
				error
			}{r, err}
		}()
	}

	go func() {
		wg.Wait()
		close(resC)
	}()

	all := search.Result{}
	for r := range resC {
		if r.error != nil {
			// Drain resC
			cancel()
			for range resC {
			}
			return nil, r.error
		}
		all.Add(r.Result)
	}

	return &all, nil
}

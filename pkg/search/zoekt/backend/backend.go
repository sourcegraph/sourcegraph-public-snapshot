// Package backend implements the built-in search providers such as indexed
// search and our JIT searcher.
package backend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/query"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

// Mock is a mock zoekt.Searcher
type Mock struct {
	Result *zoekt.Result
	Error  error

	LastQ    query.Q
	LastOpts *zoekt.Options
}

func (m *Mock) Search(ctx context.Context, q query.Q, opts *zoekt.Options) (*zoekt.Result, error) {
	m.LastQ = q
	m.LastOpts = opts
	return m.Result, m.Error
}

func (m *Mock) Close() {}

func (m *Mock) String() string { return "mock" }

type shard struct {
	zoekt.Searcher
	query.Q
	*zoekt.Options
}

type searchResponse struct {
	*zoekt.Result
	error
}

func shardedSearch(ctx context.Context, shards <-chan shard) (*zoekt.Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	resC := make(chan searchResponse)
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := shard.Searcher.Search(ctx, shard.Q, shard.Options)
			resC <- searchResponse{r, err}
		}()
	}

	go func() {
		wg.Wait()
		close(resC)
	}()

	all := zoekt.Result{}
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

func handleError(source zoekt.Source, r zoekt.Repository, err error) (*zoekt.RepositoryStatus, error) {
	status := zoekt.RepositoryStatusSearched
	if vcs.IsRepoNotExist(err) {
		if vcs.IsCloneInProgress(err) {
			status = zoekt.RepositoryStatusCloning
		} else {
			status = zoekt.RepositoryStatusMissing
		}
	} else if git.IsRevisionNotFound(err) {
		status = zoekt.RepositoryStatusCommitMissing
	} else if errcode.IsNotFound(err) {
		status = zoekt.RepositoryStatusMissing
	} else if errcode.IsTimeout(err) || errcode.IsTemporary(err) {
		status = zoekt.RepositoryStatusTimedOut
	} else if err != nil {
		return nil, err
	}
	return &zoekt.RepositoryStatus{
		Repository: r,
		Source:     source,
		Status:     status,
	}, nil
}

type semaphore chan struct{}

// Acquire increments the semaphore. Up to cap(sem) can be acquired
// concurrently. If the context is canceled before acquiring the context
// error is returned.
func (sem semaphore) Acquire(ctx context.Context) error {
	select {
	case sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release decrements the semaphore.
func (sem semaphore) Release() {
	<-sem
}

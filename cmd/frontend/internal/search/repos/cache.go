package repos

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// CachedList provides a shared cache around DefaultReposFunc. All calls to a
// List will be cached for a hardcoded duration (1min).
//
// This performance optimization exists for large instances. DefaultReposFunc
// is on the critical path for global search queries. For example on
// Sourcegraph.com it takes 3s to call the DefaultReposFunc. With this cache
// it takes a few ms.
type CachedList struct {
	mu    sync.Mutex
	cache atomic.Value
}

// Wrap fill to be cached. The wrapped function will return the cached default
// repos list. If we have a cache miss, fill is called to populate the
// cache. If the cache needs refreshing, fill will be called in a background
// goroutine and a potentially stale list will be returned.
func (s *CachedList) Wrap(fill DefaultReposFunc) DefaultReposFunc {
	return func(ctx context.Context) ([]*types.RepoName, error) {
		return s.list(ctx, fill)
	}
}

func (s *CachedList) list(ctx context.Context, fill DefaultReposFunc) ([]*types.RepoName, error) {
	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	// We don't have any repos yet, fetch them
	if len(repos) == 0 {
		return s.refreshCache(ctx, fill)
	}

	// We have existing repos, return the stale data and start background refresh
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		_, err := s.refreshCache(newCtx, fill)
		if err != nil {
			log15.Error("Refreshing default repos cache", "error", err)
		}
	}()
	return repos, nil
}

func (s *CachedList) refreshCache(ctx context.Context, fill DefaultReposFunc) ([]*types.RepoName, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check whether another routine already did the work
	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	repos, err := fill(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "querying for default repos")
	}

	s.cache.Store(&cachedRepos{
		// Copy since repos will be mutated by the caller
		repos:   append([]*types.RepoName{}, repos...),
		fetched: time.Now(),
	})

	return repos, nil
}

// defaultReposMaxAge is how long we cache the list of default repos. The list
// changes very rarely, so we can cache for a while.
const defaultReposMaxAge = time.Minute

type cachedRepos struct {
	repos   []*types.RepoName
	fetched time.Time
}

// Repos returns the current cached repos and boolean value indicating
// whether an update is required
func (c *cachedRepos) Repos() ([]*types.RepoName, bool) {
	if c == nil {
		return nil, true
	}
	if c.repos == nil {
		return nil, true
	}
	return append([]*types.RepoName{}, c.repos...), time.Since(c.fetched) > defaultReposMaxAge
}

package db

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

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

type defaultRepos struct {
	cache atomic.Value
	mu    sync.Mutex
}

func (s *defaultRepos) List(ctx context.Context) (results []*types.RepoName, err error) {
	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	// We don't have any repos yet, fetch them
	if len(repos) == 0 {
		return s.refreshCache(ctx)
	}

	// We have existing repos, return the stale data and start background refresh
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		_, err := s.refreshCache(newCtx)
		if err != nil {
			log15.Error("Refreshing default repos cache", "error", err)
		}
	}()
	return repos, nil
}

func (s *defaultRepos) refreshCache(ctx context.Context) ([]*types.RepoName, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check whether another routine already did the work
	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	repos, err := Repos.ListAllDefaultRepos(ctx)
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

func (s *defaultRepos) resetCache() {
	s.cache.Store(&cachedRepos{})
}

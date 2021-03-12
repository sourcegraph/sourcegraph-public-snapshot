package database

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type IndexableRepoStore struct {
	*basestore.Store

	cache atomic.Value
	once  sync.Once

	mu sync.Mutex
}

// IndexableRepos instantiates and returns a new IndexableRepoStore with prepared statements.
func IndexableRepos(db dbutil.DB) *IndexableRepoStore {
	return &IndexableRepoStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// IndexableReposWith instantiates and returns a new IndexableRepoStore using the other store handle.
func IndexableReposWith(other basestore.ShareableStore) *IndexableRepoStore {
	return &IndexableRepoStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *IndexableRepoStore) With(other basestore.ShareableStore) *IndexableRepoStore {
	return &IndexableRepoStore{Store: s.Store.With(other)}
}

func (s *IndexableRepoStore) Transact(ctx context.Context) (*IndexableRepoStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &IndexableRepoStore{Store: txBase}, err
}

// ensureStore instantiates a basestore.Store if necessary, using the dbconn.Global handle.
// This function ensures access to dbconn happens after the rest of the code or tests have
// initialized it.
func (s *IndexableRepoStore) ensureStore() {
	s.once.Do(func() {
		if s.Store == nil {
			s.Store = basestore.NewWithDB(dbconn.Global, sql.TxOptions{})
		}
	})
}

func (s *IndexableRepoStore) List(ctx context.Context) (results []*types.RepoName, err error) {
	s.ensureStore()

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
			log15.Error("Refreshing indexable repos cache", "error", err)
		}
	}()
	return repos, nil
}

func (s *IndexableRepoStore) refreshCache(ctx context.Context) ([]*types.RepoName, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check whether another routine already did the work
	cached, _ := s.cache.Load().(*cachedRepos)
	repos, needsUpdate := cached.Repos()
	if !needsUpdate {
		return repos, nil
	}

	repos, err := ReposWith(s).ListIndexableRepos(ctx, ListIndexableReposOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "querying for indexable repos")
	}

	s.cache.Store(&cachedRepos{
		// Copy since repos will be mutated by the caller
		repos:   append([]*types.RepoName{}, repos...),
		fetched: time.Now(),
	})

	return repos, nil
}

func (s *IndexableRepoStore) resetCache() {
	s.cache.Store(&cachedRepos{})
}

// indexableReposMaxAge is how long we cache the list of indexable repos. The list
// changes very rarely, so we can cache for a while.
const indexableReposMaxAge = time.Minute

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
	return append([]*types.RepoName{}, c.repos...), time.Since(c.fetched) > indexableReposMaxAge
}

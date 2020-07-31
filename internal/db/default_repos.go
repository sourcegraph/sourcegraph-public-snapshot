package db

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// defaultReposMaxAge is how long we cache the list of default repos. The list
// changes very rarely, so we can cache for a while.
const defaultReposMaxAge = time.Minute

type cachedRepos struct {
	repos   []*types.Repo
	fetched time.Time
}

func (c *cachedRepos) Repos() []*types.Repo {
	if c == nil || time.Since(c.fetched) > defaultReposMaxAge {
		return nil
	}
	return append([]*types.Repo{}, c.repos...)
}

type defaultRepos struct {
	mu    sync.Mutex
	cache *cachedRepos
}

func (s *defaultRepos) List(ctx context.Context) (results []*types.Repo, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cached := s.cache.Repos()
	if cached != nil {
		return cached, nil
	}
	// We continue to hold the lock here so that only one process goroutine performs the query

	const q = `
SELECT default_repos.repo_id, repo.name
FROM default_repos
JOIN repo
ON default_repos.repo_id = repo.id
`
	rows, err := dbconn.Global.QueryContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "querying default_repos table")
	}
	defer rows.Close()
	var repos []*types.Repo
	for rows.Next() {
		var r types.Repo
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, errors.Wrap(err, "scanning row from default_repos table")
		}
		repos = append(repos, &r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "scanning rows from default_repos table")
	}

	s.cache = &cachedRepos{
		// Copy since repos will be mutated by the caller
		repos:   append([]*types.Repo{}, repos...),
		fetched: time.Now(),
	}

	return repos, nil
}

func (s *defaultRepos) resetCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = &cachedRepos{}
}

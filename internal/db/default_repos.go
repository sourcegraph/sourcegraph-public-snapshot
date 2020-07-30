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

type defaultRepos struct {
	mu      sync.Mutex
	cache   []*types.Repo
	fetched time.Time
}

func (s *defaultRepos) List(ctx context.Context) (results []*types.Repo, err error) {
	s.mu.Lock()
	cached, fetched := s.cache, s.fetched
	s.mu.Unlock()

	if time.Since(fetched) < defaultReposMaxAge {
		// Return a copy since the cached slice may be mutated
		return append([]*types.Repo{}, cached...), nil
	}

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

	s.mu.Lock()
	s.cache = repos
	s.fetched = time.Now()
	s.mu.Unlock()

	// Return a copy since the cached slice may be mutated
	return append([]*types.Repo{}, repos...), nil
}

func (s *defaultRepos) resetCache() {
	s.mu.Lock()
	s.cache = nil
	s.fetched = time.Unix(0, 0)
	s.mu.Unlock()
}

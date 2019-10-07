package db

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

type defaultRepos struct{}

func (s *defaultRepos) List(ctx context.Context) (results []*types.Repo, err error) {
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
	return repos, nil
}

package db

import (
	"context"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type defaultRepos struct{}

func (s *defaultRepos) List(ctx context.Context) (results []*types.Repo, err error) {
	q := `
SELECT repo_id
FROM default_repos
`
	rows, err := dbconn.Global.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []*types.Repo
	for rows.Next() {
		var repo types.Repo
		if err := scanRepo(rows, &repo); err != nil {
			return nil, err
		}
		repos = append(repos, &repo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: This enforces repository permissions
	return authzFilter(ctx, repos, authz.Read)
}

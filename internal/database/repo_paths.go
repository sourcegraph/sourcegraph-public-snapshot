package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoPathStore interface {
	EnsureExist(ctx context.Context, repoID api.RepoID, paths []string) (int, error)
	StatsCount(context.Context, api.RepoID, string) (RepoPathStats, error)
}

type RepoPathStats struct {
	TotalFiles         int32
	AssignedOwnerFiles int32
}

type repoPaths struct {
	*basestore.Store
}

var _ RepoPathStore = &repoPaths{}

// TODO we need to precompute this.
func (r *repoPaths) StatsCount(context.Context, api.RepoID, string) (RepoPathStats, error) {
	return RepoPathStats{}, nil
}

var findPathsFmtstr = `
	WITH new_paths (absolute_path) AS (
		%s
	)
	SELECT n.absolute_path
	FROM new_paths AS n
	LEFT JOIN repo_paths AS p
	USING (absolute_path)
	WHERE p.repo_id = %s
	AND p.absolute_path IS NULL
`

const ensureExistsBatchSize = 1000

func (r *repoPaths) EnsureExist(ctx context.Context, repoID api.RepoID, paths []string) (int, error) {
	var notExist []string
	for i := 0; i < len(paths); i += ensureExistsBatchSize {
		var params []*sqlf.Query
		for _, p := range paths[i : i+ensureExistsBatchSize] {
			params = append(params, sqlf.Sprintf("SELECT %s", p))
		}
		q := sqlf.Sprintf(findPathsFmtstr, sqlf.Join(params, "UNION ALL"), repoID)
		rs, err := r.Store.Query(ctx, q)
		if err != nil {
			return 0, errors.Wrapf(err, "query: %s", q.Query(sqlf.PostgresBindVar))
		}
		for rs.Next() {
			var path string
			if err := rs.Scan(&path); err != nil {
				return 0, err
			}
			notExist = append(notExist, path)
		}
	}
	_, err := ensureRepoPaths(ctx, r.Store, notExist, repoID)
	return len(notExist), err
}

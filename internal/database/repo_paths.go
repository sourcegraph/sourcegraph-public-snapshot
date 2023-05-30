package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type RepoPathStore interface {
	EnsureExist(ctx context.Context, repoID api.RepoID, paths []string) error
}

type repoPaths struct {
	*basestore.Store
}

var _ RepoPathStore = &repoPaths{}

var findPathsFmtstr = `
	SELECT absolute_path
	FROM repo_paths
	WHERE absolute_path IN (%s)
`

const ensureExistsBatchSize = 1000

func (r *repoPaths) EnsureExist(ctx context.Context, repoID api.RepoID, paths []string) error {
	// 1. Find which of the given paths do not exist in repo_paths table:
	var notExist []string
	for i := 0; i < len(paths); i += ensureExistsBatchSize {
		var params []*sqlf.Query
		for _, p := range paths {
			params = append(params, sqlf.Sprintf("%s", p))
		}
		q := sqlf.Sprintf(findPathsFmtstr, sqlf.Join(params, ","))
		rs, err := r.Store.Query(ctx, q)
		if err != nil {
			return err
		}
		for rs.Next() {
			var path string
			if err := rs.Scan(&path); err != nil {
				return err
			}
			notExist = append(notExist, path)
		}
	}
	_, err := ensureRepoPaths(ctx, r.Store, notExist, repoID)
	return err
}

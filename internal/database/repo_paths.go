package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoPathStore interface {
	EnsureExist(ctx context.Context, repoID api.RepoID, paths []string) (map[string]int, error)
	UpdateCounts(ctx context.Context, repoID api.RepoID, deepCounts FileWalkable[int]) (int, error)
}

type FileWalkable[T any] interface {
	Walk(func(path string, value T) error) error
}

type repoPaths struct {
	*basestore.Store
}

var _ RepoPathStore = &repoPaths{}

var findPathsFmtstr = `
	WITH new_paths (absolute_path) AS (
		%s
	)
	SELECT n.absolute_path, COALESCE(p.id, 0)
	FROM new_paths AS n
	LEFT JOIN repo_paths AS p
	USING (absolute_path)
	WHERE p.repo_id IS NULL OR p.repo_id = %s
`

const ensureExistsBatchSize = 1000

// TODO: Delete old paths
func (r *repoPaths) EnsureExist(ctx context.Context, repoID api.RepoID, paths []string) (map[string]int, error) {
	ids := map[string]int{}
	var notExist []string
	for i := 0; i < len(paths); i += ensureExistsBatchSize {
		var params []*sqlf.Query
		top := i + ensureExistsBatchSize
		if max := len(paths); top > max {
			top = max
		}
		for _, p := range paths[i:top] {
			params = append(params, sqlf.Sprintf("SELECT %s", p))
		}
		q := sqlf.Sprintf(findPathsFmtstr, sqlf.Join(params, "UNION ALL"), repoID)
		rs, err := r.Store.Query(ctx, q)
		if err != nil {
			return nil, errors.Wrapf(err, "query: %s", q.Query(sqlf.PostgresBindVar))
		}
		for rs.Next() {
			var path string
			var id int
			if err := rs.Scan(&path, &id); err != nil {
				return nil, err
			}
			if id == 0 {
				notExist = append(notExist, path)
			} else {
				ids[path] = id
			}
		}
	}
	newIDs, err := ensureRepoPaths(ctx, r.Store, notExist, repoID)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(newIDs); i++ {
		ids[notExist[i]] = newIDs[i]
	}
	return ids, err
}

func (r *repoPaths) UpdateCounts(ctx context.Context, repoID api.RepoID, deepCounts FileWalkable[int]) (int, error) {
	var updatedRows int
	err := deepCounts.Walk(func(path string, count int) error {
		q := sqlf.Sprintf("UPDATE repo_paths SET deep_file_count = %s WHERE repo_id = %s AND absolute_path = %s RETURNING id", count, repoID, path)
		r := r.Store.QueryRow(ctx, q)
		var resultID int
		if err := r.Scan(&resultID); err != nil {
			return errors.Wrapf(err, "query: %s, repo: %d, path: %s", q.Query(sqlf.PostgresBindVar), repoID, path)
		}
		if resultID > 0 {
			updatedRows++
		}
		return nil
	})
	return updatedRows, err
}

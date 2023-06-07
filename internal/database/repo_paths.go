package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RepoTreeCounts allows iterating over file paths and yield total counts
// of all the files within a file tree rooted at given path.
type RepoTreeCounts interface {
	Iterate(func(path string, totalFiles int) error) error
}

type RepoPathStore interface {
	// UpdateFileCounts inserts file counts for every iterated path at given repository.
	// If any of the iterated paths does not exist, it's created. Returns the number of updated paths.
	UpdateFileCounts(context.Context, api.RepoID, RepoTreeCounts, time.Time) (int, error)
	// AggregateFileCount returns the file count aggregated for given TreeLocationOps.
	// For instance, TreeLocationOpts with RepoID and Path returns counts for tree at given path,
	// setting only RepoID gives counts for repo root, while setting none gives counts for the whole
	// instance. Lack of data counts as 0.
	AggregateFileCount(context.Context, TreeLocationOpts) (int32, error)
}

var _ RepoPathStore = &repoPathStore{}

type repoPathStore struct {
	*basestore.Store
}

var updateFileCountsFmtstr = `
	UPDATE repo_paths
	SET tree_files_count = %s,
	tree_files_counts_updated_at = %s
	WHERE id = %s
`

func (s *repoPathStore) UpdateFileCounts(ctx context.Context, repoID api.RepoID, counts RepoTreeCounts, timestamp time.Time) (int, error) {
	var rowsUpdated int
	err := counts.Iterate(func(path string, totalFiles int) error {
		pathIDs, err := ensureRepoPaths(ctx, s.Store, []string{path}, repoID)
		if err != nil {
			return err
		}
		if got, want := len(pathIDs), 1; got != want {
			return errors.Newf("want exactly 1 repo path, got %d", got)
		}
		res, err := s.ExecResult(ctx, sqlf.Sprintf(updateFileCountsFmtstr, totalFiles, timestamp, pathIDs[0]))
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		rowsUpdated += int(rows)
		return nil
	})
	return rowsUpdated, err
}

var aggregateFileCountFmtstr = `
    SELECT SUM(COALESCE(p.tree_files_count, 0))
    FROM repo_paths AS p
    WHERE p.absolute_path = %s
`

func (s *repoPathStore) AggregateFileCount(ctx context.Context, opts TreeLocationOpts) (int32, error) {
	var qs []*sqlf.Query
	qs = append(qs, sqlf.Sprintf(aggregateFileCountFmtstr, opts.Path))
	if repoID := opts.RepoID; repoID != 0 {
		qs = append(qs, sqlf.Sprintf("AND p.repo_id = %s", repoID))
	}
	var count int32
	if err := s.Store.QueryRow(ctx, sqlf.Join(qs, "\n")).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

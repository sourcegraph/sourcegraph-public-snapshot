package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CodeownedTreeWalk interface {
	Walk(func(path string, ownerCounts []CodeownedTreeCounts) error) error
}

type CodeownedTreeCounts struct {
	Reference string
	FileCount int
}

type OwnershipStatsStore interface {
	// UpdateIndividualCounts walks a representation of a repo file tree
	// that yields ownership information for each file and directory, and persists
	// that in the database.
	UpdateIndividualCounts(ctx context.Context, repoID api.RepoID, data CodeownedTreeWalk, timestamp time.Time) (int, error)
}

var _ OwnershipStatsStore = &ownershipStats{}

type ownershipStats struct {
	*basestore.Store
}

var codeownerQueryFmtstr = `
	WITH existing (id) AS (
		SELECT a.id
		FROM codeowners_owners AS a
		WHERE a.reference = %s
	), inserted (id) AS (
		INSERT INTO codeowners_owners (reference)
		SELECT %s
		WHERE NOT EXISTS (SELECT id FROM existing)
		RETURNING id
	)
	SELECT id FROM existing
	UNION ALL
	SELECT id FROM inserted
`

var codeownerUpsertCountsFmtstr = `
	INSERT INTO codeowners_individual_stats (file_path_id, owner_id, tree_owned_files_count, last_updated_at)
	SELECT p.id, %s, %s, %s
	FROM repo_paths AS p
	WHERE p.repo_id = %s
	AND p.absolute_path = %s
	ON CONFLICT (file_path_id, owner_id)
	DO UPDATE SET
		tree_owned_files_count = EXCLUDED.tree_owned_files_count,
		last_updated_at = EXCLUDED.last_updated_at
`

func (s *ownershipStats) UpdateIndividualCounts(ctx context.Context, repoID api.RepoID, data CodeownedTreeWalk, timestamp time.Time) (int, error) {
	codeownersCache := map[string]int{} // Cache codeowner ID by reference
	var totalRows int
	err := data.Walk(func(path string, ownerCounts []CodeownedTreeCounts) error {
		for _, owner := range ownerCounts {
			id := codeownersCache[owner.Reference]
			if id == 0 {
				q := sqlf.Sprintf(codeownerQueryFmtstr, owner.Reference, owner.Reference)
				r := s.Store.QueryRow(ctx, q)
				if err := r.Scan(&id); err != nil {
					return errors.Wrapf(err, "querying/adding owner %q failed, query: %s", owner.Reference, q.Query(sqlf.PostgresBindVar))
				}
				codeownersCache[owner.Reference] = id
			}
			// At this point we assume paths exists in repo_paths, otherwise we will not update.
			q := sqlf.Sprintf(codeownerUpsertCountsFmtstr, id, owner.FileCount, timestamp, repoID, path)
			res, err := s.Store.ExecResult(ctx, q)
			if err != nil {
				return errors.Wrapf(err, "updating counts for %q at repoID=%d path=%s failed, query: %s", owner.Reference, repoID, path, q.Query(sqlf.PostgresBindVar))
			}
			rows, err := res.RowsAffected()
			if err != nil {
				return errors.Wrapf(err, "updating counts for %q at repoID=%d path=%s failed, query: %s", owner.Reference, repoID, path, q.Query(sqlf.PostgresBindVar))
			}
			totalRows += int(rows)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalRows, nil
}

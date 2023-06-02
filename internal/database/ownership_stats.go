package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FileOwnershipAggregate allows iterating through the file tree
// of a repository, providing ownership counts for every owner
// and every directory.
type FileOwnershipAggregate interface {
	Iterate(func(path string, counts TreeCounts) error) error
}

// TreeCounts describes ownership magnitude by file count for given owner.
// The scope of ownership is contextual, and can range from a file tree
// in case of FileOwnershipAggregate to whole instance when querying
// without restrictions through QueryIndividualCounts.
type TreeCounts struct {
	// CodeownersReference is the text found in CODEOWNERS files that matched the counted files in this file tree.
	CodeownersReference string
	// CodeownedFileCount is the number of files that matched given owner in this file tree
	CodeownedFileCount int
}

// TreeLocationOpts allows locating and aggregating statistics on file trees.
type TreeLocationOpts struct {
	// RepoID locates a file tree for given repo.
	// If 0 then all repos all considered.
	RepoID api.RepoID

	// Path locates a file tree within a given repo.
	// Empty path "" represents repo root.
	// Paths do not contain leading /.
	Path string
}

type OwnershipStatsStore interface {
	// UpdateIndividualCounts walks a representation of a repo file tree
	// that yields ownership information for each file and directory, and persists
	// that in the database. All the counts are marked by given update timestamp.
	UpdateIndividualCounts(context.Context, api.RepoID, FileOwnershipAggregate, time.Time) (int, error)

	// QueryIndividualCounts looks up and aggregates data for individual stats of located file trees.
	// To find ownership for the whole instance, use empty TreeLocationOpts.
	// To find ownership for the repo root, only specify RepoID in TreeLocationOpts.
	// To find ownership for specific file tree, specify RepoID and Path in TreeLocationOpts.
	QueryIndividualCounts(context.Context, TreeLocationOpts, *LimitOffset) ([]TreeCounts, error)
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

// TODO: Introduce batch inserter.
func (s *ownershipStats) UpdateIndividualCounts(ctx context.Context, repoID api.RepoID, data FileOwnershipAggregate, timestamp time.Time) (int, error) {
	codeownersCache := map[string]int{} // Cache codeowner ID by reference
	var totalRows int
	err := data.Iterate(func(path string, counts TreeCounts) error {
		id := codeownersCache[counts.CodeownersReference]
		if id == 0 {
			q := sqlf.Sprintf(codeownerQueryFmtstr, counts.CodeownersReference, counts.CodeownersReference)
			r := s.Store.QueryRow(ctx, q)
			if err := r.Scan(&id); err != nil {
				return errors.Wrapf(err, "querying/adding owner %q failed, query: %s", counts.CodeownersReference, q.Query(sqlf.PostgresBindVar))
			}
			codeownersCache[counts.CodeownersReference] = id
		}
		// At this point we assume paths exists in repo_paths, otherwise we will not update.
		q := sqlf.Sprintf(codeownerUpsertCountsFmtstr, id, counts.CodeownedFileCount, timestamp, repoID, path)
		res, err := s.Store.ExecResult(ctx, q)
		if err != nil {
			return errors.Wrapf(err, "updating counts for %q at repoID=%d path=%s failed, query: %s", counts.CodeownersReference, repoID, path, q.Query(sqlf.PostgresBindVar))
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return errors.Wrapf(err, "updating counts for %q at repoID=%d path=%s failed, query: %s", counts.CodeownersReference, repoID, path, q.Query(sqlf.PostgresBindVar))
		}
		totalRows += int(rows)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalRows, nil
}

var aggregateOwnershipFmtstr = `
	SELECT o.reference, SUM(s.tree_owned_files_count)
	FROM codeowners_individual_stats AS s
	INNER JOIN repo_paths AS p ON s.file_path_id = p.id
	INNER JOIN codeowners_owners AS o ON o.id = s.owner_id
	WHERE p.absolute_path = %s
`

func (s *ownershipStats) QueryIndividualCounts(ctx context.Context, opts TreeLocationOpts, limitOffset *LimitOffset) ([]TreeCounts, error) {
	qs := []*sqlf.Query{sqlf.Sprintf(aggregateOwnershipFmtstr, opts.Path)}
	if repoID := opts.RepoID; repoID != 0 {
		qs = append(qs, sqlf.Sprintf("AND p.repo_id = %s", repoID))
	}
	qs = append(qs, sqlf.Sprintf("GROUP BY 1 ORDER BY 2 DESC, 1 ASC"))
	qs = append(qs, limitOffset.SQL())
	rs, err := s.Store.Query(ctx, sqlf.Join(qs, "\n"))
	if err != nil {
		return nil, err
	}
	var owners []TreeCounts
	for rs.Next() {
		var o TreeCounts
		if err := rs.Scan(&o.CodeownersReference, &o.CodeownedFileCount); err != nil {
			return nil, err
		}
		owners = append(owners, o)
	}
	return owners, nil
}

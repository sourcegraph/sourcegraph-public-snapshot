package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TreeCodeownersStats allows iterating through the file tree
// of a repository, providing ownership counts for every owner
// and every directory.
type TreeCodeownersStats interface {
	Iterate(func(path string, counts PathCodeownersCounts) error) error
}

// PathCodeownersCounts describes ownership magnitude by file count for given owner.
// The scope of ownership is contextual, and can range from a file tree
// in case of TreeCodeownersStats to whole instance when querying
// without restrictions through QueryIndividualCounts.
type PathCodeownersCounts struct {
	// CodeownersReference is the text found in CODEOWNERS files that matched the counted files in this file tree.
	CodeownersReference string
	// CodeownedFileCount is the number of files that matched given owner in this file tree.
	CodeownedFileCount int
}

// TreeAggregateStats allows iterating through the file tree of a repository
// providing ownership data that is aggregated by file path only (as opposed
// to TreeCodeownersStats)
type TreeAggregateStats interface {
	Iterate(func(path string, counts PathAggregateCounts) error) error
}

type PathAggregateCounts struct {
	// CodeownedFileCount is the total number of files nested within given tree root
	// that are owned via CODEOWNERS.
	CodeownedFileCount int
	// AssignedOwnershipFileCount is the total number of files in tree that are owned via assigned ownership.
	AssignedOwnershipFileCount int
	// TotalOwnedFileCount is the total number of files in tree that have any ownership associated
	// - either via CODEOWNERS or via assigned ownership.
	TotalOwnedFileCount int
	// UpdatedAt shows When statistics were last updated.
	UpdatedAt time.Time
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
	// UpdateIndividualCounts iterates given data about individual CODEOWNERS ownership
	// and persists it in the database. All the counts are marked by given update timestamp.
	UpdateIndividualCounts(context.Context, api.RepoID, TreeCodeownersStats, time.Time) (int, error)

	// UpdateAggregateCounts iterates given data about aggregate ownership over
	// a given file tree, and persists it in the database. All the counts are marked
	// by given update timestamp.
	UpdateAggregateCounts(context.Context, api.RepoID, TreeAggregateStats, time.Time) (int, error)

	// QueryIndividualCounts looks up and aggregates data for individual stats of located file trees.
	// To find ownership for the whole instance, use empty TreeLocationOpts.
	// To find ownership for the repo root, only specify RepoID in TreeLocationOpts.
	// To find ownership for specific file tree, specify RepoID and Path in TreeLocationOpts.
	QueryIndividualCounts(context.Context, TreeLocationOpts, *LimitOffset) ([]PathCodeownersCounts, error)

	// QueryAggregateCounts looks up ownership aggregate data for a file tree. At
	// this point these include total count of files that are owned via CODEOWNERS
	// and assigned ownership.
	QueryAggregateCounts(context.Context, TreeLocationOpts) (PathAggregateCounts, error)
}

var _ OwnershipStatsStore = &ownershipStats{}

type ownershipStats struct {
	*basestore.Store
}

const codeownerQueryFmtstr = `
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

const codeownerUpsertCountsFmtstr = `
	INSERT INTO codeowners_individual_stats (file_path_id, owner_id, tree_owned_files_count, updated_at)
	VALUES (%s, %s, %s, %s)
	ON CONFLICT (file_path_id, owner_id)
	DO UPDATE SET
		tree_owned_files_count = EXCLUDED.tree_owned_files_count,
		updated_at = EXCLUDED.updated_at
`

func (s *ownershipStats) UpdateIndividualCounts(ctx context.Context, repoID api.RepoID, data TreeCodeownersStats, timestamp time.Time) (int, error) {
	codeownersCache := map[string]int{} // Cache codeowner ID by reference
	var totalRows int
	err := data.Iterate(func(path string, counts PathCodeownersCounts) error {
		ownerID := codeownersCache[counts.CodeownersReference]
		if ownerID == 0 {
			q := sqlf.Sprintf(codeownerQueryFmtstr, counts.CodeownersReference, counts.CodeownersReference)
			r := s.Store.QueryRow(ctx, q)
			if err := r.Scan(&ownerID); err != nil {
				return errors.Wrapf(err, "querying/adding owner %q failed", counts.CodeownersReference)
			}
			codeownersCache[counts.CodeownersReference] = ownerID
		}
		pathIDs, err := ensureRepoPaths(ctx, s.Store, []string{path}, repoID)
		if err != nil {
			return err
		}
		if got, want := len(pathIDs), 1; got != want {
			return errors.Newf("want exactly 1 repo path, got %d", got)
		}
		// At this point we assume paths exists in repo_paths, otherwise we will not update.
		q := sqlf.Sprintf(codeownerUpsertCountsFmtstr, pathIDs[0], ownerID, counts.CodeownedFileCount, timestamp)
		res, err := s.Store.ExecResult(ctx, q)
		if err != nil {
			return errors.Wrapf(err, "updating counts for %q at repoID=%d path=%s failed", counts.CodeownersReference, repoID, path)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return errors.Wrapf(err, "updating counts for %q at repoID=%d path=%s failed", counts.CodeownersReference, repoID, path)
		}
		totalRows += int(rows)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalRows, nil
}

const aggregateCountsUpdateFmtstr = `
	INSERT INTO ownership_path_stats (
		file_path_id,
		tree_codeowned_files_count,
		tree_assigned_ownership_files_count,
		tree_any_ownership_files_count,
		last_updated_at)
	VALUES (%s, %s, %s, %s, %s)
	ON CONFLICT (file_path_id)
	DO UPDATE SET
	tree_codeowned_files_count = EXCLUDED.tree_codeowned_files_count,
	tree_assigned_ownership_files_count = EXCLUDED.tree_assigned_ownership_files_count,
	tree_any_ownership_files_count = EXCLUDED.tree_any_ownership_files_count,
	last_updated_at = EXCLUDED.last_updated_at
`

func (s *ownershipStats) UpdateAggregateCounts(ctx context.Context, repoID api.RepoID, data TreeAggregateStats, timestamp time.Time) (int, error) {
	var totalUpdates int
	err := data.Iterate(func(path string, counts PathAggregateCounts) error {
		pathIDs, err := ensureRepoPaths(ctx, s.Store, []string{path}, repoID)
		if err != nil {
			return err
		}
		if got, want := len(pathIDs), 1; got != want {
			return errors.Newf("want exactly 1 repo path, got %d", got)
		}
		q := sqlf.Sprintf(
			aggregateCountsUpdateFmtstr,
			pathIDs[0],
			counts.CodeownedFileCount,
			counts.AssignedOwnershipFileCount,
			counts.TotalOwnedFileCount,
			timestamp,
		)
		res, err := s.ExecResult(ctx, q)
		if err != nil {
			return errors.Wrapf(err, "updating counts at repoID=%d path=%s failed", repoID, path)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return errors.Wrapf(err, "getting result of updating counts at repoID=%d path=%s failed", repoID, path)
		}
		totalUpdates += int(rows)
		return nil
	})
	return totalUpdates, err
}

const aggregateOwnershipFmtstr = `
	SELECT o.reference, SUM(COALESCE(s.tree_owned_files_count, 0))
	FROM codeowners_individual_stats AS s
	INNER JOIN repo_paths AS p ON s.file_path_id = p.id
	INNER JOIN codeowners_owners AS o ON o.id = s.owner_id
	WHERE p.absolute_path = %s
`

var treeCountsScanner = basestore.NewSliceScanner(func(s dbutil.Scanner) (PathCodeownersCounts, error) {
	var cs PathCodeownersCounts
	err := s.Scan(&cs.CodeownersReference, &cs.CodeownedFileCount)
	return cs, err
})

func (s *ownershipStats) QueryIndividualCounts(ctx context.Context, opts TreeLocationOpts, limitOffset *LimitOffset) ([]PathCodeownersCounts, error) {
	qs := []*sqlf.Query{sqlf.Sprintf(aggregateOwnershipFmtstr, opts.Path)}
	if repoID := opts.RepoID; repoID != 0 {
		qs = append(qs, sqlf.Sprintf("AND p.repo_id = %s", repoID))
	}
	qs = append(qs, sqlf.Sprintf("GROUP BY 1 ORDER BY 2 DESC, 1 ASC"))
	qs = append(qs, limitOffset.SQL())
	return treeCountsScanner(s.Store.Query(ctx, sqlf.Join(qs, "\n")))
}

const treeAggregateCountsFmtstr = `
	WITH signal_config AS (SELECT * FROM own_signal_configurations WHERE name = 'analytics' LIMIT 1)
	SELECT
		SUM(COALESCE(s.tree_codeowned_files_count, 0)),
		SUM(COALESCE(s.tree_assigned_ownership_files_count, 0)),
		SUM(COALESCE(s.tree_any_ownership_files_count, 0)),
		MAX(s.last_updated_at)
	FROM ownership_path_stats AS s
	INNER JOIN repo_paths AS p ON s.file_path_id = p.id
	WHERE p.absolute_path = %s AND p.repo_id NOT IN (SELECT repo.id FROM repo, signal_config WHERE repo.name ~~ ANY(signal_config.excluded_repo_patterns))
`

func (s *ownershipStats) QueryAggregateCounts(ctx context.Context, opts TreeLocationOpts) (PathAggregateCounts, error) {
	qs := []*sqlf.Query{sqlf.Sprintf(treeAggregateCountsFmtstr, opts.Path)}
	if repoID := opts.RepoID; repoID != 0 {
		qs = append(qs, sqlf.Sprintf("AND p.repo_id = %s", repoID))
	}
	var cs PathAggregateCounts
	err := s.Store.QueryRow(ctx, sqlf.Join(qs, "\n")).Scan(
		&dbutil.NullInt{N: &cs.CodeownedFileCount},
		&dbutil.NullInt{N: &cs.AssignedOwnershipFileCount},
		&dbutil.NullInt{N: &cs.TotalOwnedFileCount},
		&dbutil.NullTime{Time: &cs.UpdatedAt},
	)
	return cs, err
}

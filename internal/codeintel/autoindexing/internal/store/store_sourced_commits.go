package store

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetStaleSourcedCommits returns a set of commits attached to repositories that have been
// least recently checked for resolvability via gitserver. We do this periodically in
// order to determine which records in the database are unreachable by normal query
// paths and clean up that occupied (but useless) space. The output is of this method is
// ordered by repository ID then by commit.
func (s *store) GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, trace, endObservation := s.operations.getStaleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	now = now.UTC()
	interval := int(minimumTimeSinceLastCheck / time.Second)

	candidatesSubquery := sqlf.Sprintf(candidatesSubqueryCTE, now, interval)
	staleCommitsQuery := sqlf.Sprintf(staleIndexSourcedCommitsQuery, candidatesSubquery, limit)

	sourcedCommits, err := scanSourcedCommits(tx.Query(ctx, staleCommitsQuery))
	if err != nil {
		return nil, err
	}

	numCommits := 0
	for _, commits := range sourcedCommits {
		numCommits += len(commits.Commits)
	}
	trace.Log(
		log.Int("numRepositories", len(sourcedCommits)),
		log.Int("numCommits", numCommits),
	)

	return sourcedCommits, nil
}

const staleIndexSourcedCommitsQuery = `
-- source: internal/codeintel/autoindexing/internal/store/store_sourced_commits.go:StaleSourcedCommits
WITH
	candidates AS (%s)
SELECT r.id, r.name, c.commit
FROM candidates c
JOIN repo r ON r.id = c.repository_id
-- Order results so that the repositories with the commits that have been updated
-- the least frequently come first. Once a number of commits are processed from a
-- given repository the ordering may change.
ORDER BY MIN(c.max_last_checked_at) OVER (PARTITION BY c.repository_id), c.commit
LIMIT %s
`

const candidatesSubqueryCTE = `
SELECT
	repository_id,
	commit,
	-- Keep track of the most recent update of this commit that we know about
	-- as any earlier dates for the same repository and commit pair carry no
	-- useful information.
	MAX(commit_last_checked_at) as max_last_checked_at
FROM lsif_indexes
WHERE
	-- Ignore records already marked as deleted
	state NOT IN ('deleted', 'deleting') AND
	-- Ignore records that have been checked recently. Note this condition is
	-- true for a null commit_last_checked_at (which has never been checked).
	(%s - commit_last_checked_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE
GROUP BY repository_id, commit
`

// UpdateSourcedCommits updates the commit_last_checked_at field of each upload and index records belonging
// to the given repository identifier and commit. This method returns the count of upload and index records
// modified, respectively.
func (s *store) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error) {
	ctx, trace, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	candidateIndexesSubquery := sqlf.Sprintf(candidateIndexesCTE, repositoryID, commit)
	updateSourcedCommitsQuery := sqlf.Sprintf(updateSourcedCommitsQuery, candidateIndexesSubquery, now)

	indexesUpdated, err = scanCount(s.db.Query(ctx, updateSourcedCommitsQuery))
	if err != nil {
		return 0, err
	}
	trace.Log(log.Int("indexesUpdated", indexesUpdated))

	return indexesUpdated, nil
}

const updateSourcedCommitsQuery = `
-- source: internal/codeintel/autoindexing/internal/store/store_sourced_commits.go:UpdateSourcedCommits
WITH
candidate_indexes AS (%s),
update_indexes AS (
	UPDATE lsif_indexes u
	SET commit_last_checked_at = %s
	WHERE id IN (SELECT id FROM candidate_indexes)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM update_indexes) AS num_indexes
`

const candidateIndexesCTE = `
SELECT u.id
FROM lsif_indexes u
WHERE u.repository_id = %s AND u.commit = %s

-- Lock these rows in a deterministic order so that we don't
-- deadlock with other processes updating the lsif_indexes table.
ORDER BY u.id FOR UPDATE
`

// DeleteSourcedCommits deletes each index records belonging to the given repository identifier
// and commit. Indexes are hard-deleted. This method returns the count of index records modified.
//
// If a maximum commit lag is supplied, then any upload records in the uploading, queued, or processing states
// younger than the provided lag will not be deleted, but its timestamp will be modified as if the sibling method
// UpdateSourcedCommits was called instead. This configurable parameter enables support for remote code hosts
// that are not the source of truth; if we deleted all pending records without resolvable commits introduce races
// between the customer's Sourcegraph instance and their CI (and their CI will usually win).
func (s *store) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration) (indexesDeleted int, err error) {
	ctx, trace, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	unset, _ := s.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with unknown commit")
	defer unset(ctx)

	candidateIndexesSubquery := sqlf.Sprintf(candidateIndexesCTE, repositoryID, commit)
	deleteSourcedCommitsQuery := sqlf.Sprintf(deleteSourcedCommitsQuery, candidateIndexesSubquery)

	indexesDeleted, err = scanCount(s.db.Query(ctx, deleteSourcedCommitsQuery))
	if err != nil {
		return 0, err
	}
	trace.Log(log.Int("indexesDeleted", indexesDeleted))

	return indexesDeleted, nil
}

const deleteSourcedCommitsQuery = `
-- source: internal/codeintel/autoindexing/internal/store/store_sourced_commits.go:DeleteSourcedCommits
WITH
candidate_indexes AS (%s),
delete_indexes AS (
	DELETE FROM lsif_indexes u
	WHERE id IN (SELECT id FROM candidate_indexes)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM delete_indexes) AS num_indexes_deleted
`

// scanCounts scans pairs of id/counts from the return value of `*Store.query`.
func scanCounts(rows *sql.Rows, queryErr error) (_ map[int]int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}

		visibilities[id] = count
	}

	return visibilities, nil
}

// scanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func scanSourcedCommits(rows *sql.Rows, queryErr error) (_ []shared.SourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]shared.SourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = shared.SourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]shared.SourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}

func scanCount(rows *sql.Rows, queryErr error) (value int, err error) {
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return 0, err
		}
	}

	return value, nil
}

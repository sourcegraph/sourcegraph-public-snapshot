package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) ProcessStaleSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	_ time.Duration,
	shouldDelete func(ctx context.Context, repositoryID int, commit string) (bool, error),
) (int, error) {
	return s.processStaleSourcedCommits(ctx, minimumTimeSinceLastCheck, commitResolverBatchSize, shouldDelete, time.Now())
}

func (s *store) processStaleSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	shouldDelete func(ctx context.Context, repositoryID int, commit string) (bool, error),
	now time.Time,
) (totalDeleted int, err error) {
	ctx, _, endObservation := s.operations.processStaleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	now = now.UTC()
	interval := int(minimumTimeSinceLastCheck / time.Second)

	staleIndexes, err := scanSourcedCommits(tx.Query(ctx, sqlf.Sprintf(
		staleIndexSourcedCommitsQuery,
		now,
		interval,
		commitResolverBatchSize,
	)))
	if err != nil {
		return 0, err
	}

	for _, sc := range staleIndexes {
		var (
			keep   []string
			remove []string
		)

		for _, commit := range sc.Commits {
			if ok, err := shouldDelete(ctx, sc.RepositoryID, commit); err != nil {
				return 0, err
			} else if ok {
				remove = append(remove, commit)
			} else {
				keep = append(keep, commit)
			}
		}

		unset, _ := tx.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload associated with unknown commit")
		defer unset(ctx)

		indexesDeleted, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
			updateSourcedCommitsQuery,
			sc.RepositoryID,
			pq.Array(keep),
			pq.Array(remove),
			now,
			pq.Array(keep),
			pq.Array(remove),
		)))
		if err != nil {
			return 0, err
		}

		totalDeleted += indexesDeleted
	}

	return totalDeleted, nil
}

const staleIndexSourcedCommitsQuery = `
WITH candidates AS (
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
)
SELECT r.id, r.name, c.commit
FROM candidates c
JOIN repo r ON r.id = c.repository_id
-- Order results so that the repositories with the commits that have been updated
-- the least frequently come first. Once a number of commits are processed from a
-- given repository the ordering may change.
ORDER BY MIN(c.max_last_checked_at) OVER (PARTITION BY c.repository_id), c.commit
LIMIT %s
`

const updateSourcedCommitsQuery = `
WITH
candidate_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		(
			u.commit = ANY(%s) OR
			u.commit = ANY(%s)
		)

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_indexes table.
	ORDER BY u.id FOR UPDATE
),
update_indexes AS (
	UPDATE lsif_indexes
	SET commit_last_checked_at = %s
	WHERE id IN (SELECT id FROM candidate_indexes WHERE commit = ANY(%s))
	RETURNING 1
),
delete_indexes AS (
	DELETE FROM lsif_indexes
	WHERE id IN (SELECT id FROM candidate_indexes WHERE commit = ANY(%s))
	RETURNING 1
)
SELECT COUNT(*) FROM delete_indexes
`

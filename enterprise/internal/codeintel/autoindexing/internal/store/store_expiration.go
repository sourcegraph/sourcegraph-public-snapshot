package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ExpireFailedRecords removes autoindexing job records that meet the following conditions:
//
//   - The record is in the "failed" state
//   - The time between the job finishing and the current timestamp exceeds the given max age
//   - It is not the most recent-to-finish failure for the same repo, root, and indexer values
//     **unless** there is a more recent success.
func (s *store) ExpireFailedRecords(ctx context.Context, batchSize int, failedIndexMaxAge time.Duration, now time.Time) (err error) {
	ctx, _, endObservation := s.operations.expireFailedRecords.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(expireFailedRecordsQuery, now, int(failedIndexMaxAge/time.Second), batchSize))
}

const expireFailedRecordsQuery = `
WITH
ranked_indexes AS (
	SELECT
		u.*,
		RANK() OVER (
			PARTITION BY
				repository_id,
				root,
				indexer
			ORDER BY
				finished_at DESC
		) AS rank
	FROM lsif_indexes u
	WHERE
		u.state = 'failed' AND
		%s - u.finished_at >= %s * interval '1 second'
),
locked_indexes AS (
	SELECT i.id
	FROM lsif_indexes i
	JOIN ranked_indexes ri ON ri.id = i.id

	-- We either select ranked indexes that have a rank > 1, meaning
	-- there's another more recent failure in this "pipeline" that has
	-- relevant information to debug the failure.
	--
	-- If we have rank = 1, but there's a newer SUCCESSFUL record for
	-- the same "pipeline", then we can say that this failure information
	-- is no longer relevant.

	WHERE ri.rank != 1 OR EXISTS (
		SELECT 1
		FROM lsif_indexes i2
		WHERE
			i2.state = 'completed' AND
			i2.finished_at > i.finished_at AND
			i2.repository_id = i.repository_id AND
			i2.root = i.root AND
			i2.indexer = i.indexer
	)
	ORDER BY i.id
	FOR UPDATE SKIP LOCKED
	LIMIT %d
)
DELETE FROM lsif_indexes
WHERE id IN (SELECT id FROM locked_indexes)
`

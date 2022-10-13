package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) ExpireFailedRecords(ctx context.Context, failedIndexMaxAge time.Duration, now time.Time) (err error) {
	ctx, _, endObservation := s.operations.expireFailedRecords.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return (s.db.Exec(ctx, sqlf.Sprintf(expireFailedRecordsQuery, now, int(failedIndexMaxAge/time.Second))))
}

const expireFailedRecordsQuery = `
WITH
ranked_indexes AS (
	SELECT
		u.*,
		RANK() OVER (PARTITION BY repository_id, root, indexer ORDER BY finished_at DESC) AS rank
	FROM lsif_indexes u
	WHERE
		u.state = 'failed' AND
		%s - u.finished_at >= %s * interval '1 second'
),
locked_indexes AS (
	SELECT id
	FROM ranked_indexes
	WHERE rank != 1
	ORDER BY id
	FOR UPDATE SKIP LOCKED
)
DELETE FROM lsif_indexes
WHERE id IN (SELECT id FROM locked_indexes)
`

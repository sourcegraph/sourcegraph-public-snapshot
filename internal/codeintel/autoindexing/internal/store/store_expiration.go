package store

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) ExpireFailedRecords(ctx context.Context, failedIndexMaxAge time.Duration, now time.Time) (err error) {
	ctx, _, endObservation := s.operations.expireFailedRecords.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(expireFailedRecordsQuery, now, int(failedIndexMaxAge/time.Second))
	fmt.Printf("QUERY: %s %v\n", q.Query(sqlf.PostgresBindVar), q.Args())
	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, q))
	fmt.Printf("COUNT: %d\n", count)
	return err
}

const expireFailedRecordsQuery = `
WITH ranked_indexes AS (
	SELECT
		u.*,
		RANK() OVER (PARTITION BY repository_id, root, indexer ORDER BY finished_at DESC) AS rank
	FROM lsif_indexes u
	WHERE
		u.state = 'failed' AND
		%s - u.finished_at >= %s * interval '1 second'
)
SELECT COUNT(*) FROM ranked_indexes
`

package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// SetRepositoriesForRetentionScan returns a set of repository identifiers with live code intelligence
// data and a fresh associated commit graph. Repositories that were returned previously from this call
// within the  given process delay are not returned.
func (s *store) SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error) {
	ctx, _, endObservation := s.operations.setRepositoriesForRetentionScan.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	now := timeutil.Now()

	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(
		repositoryIDsForRetentionScanQuery,
		now,
		int(processDelay/time.Second),
		limit,
		now,
		now,
	)))
}

const repositoryIDsForRetentionScanQuery = `
-- source: internal/codeintel/uploads/internal/store/store_repositories.go:setRepositoriesForRetentionScan
WITH candidate_repositories AS (
	SELECT DISTINCT u.repository_id AS id
	FROM lsif_uploads u
	WHERE u.state = 'completed'
),
repositories AS (
	SELECT cr.id
	FROM candidate_repositories cr
	LEFT JOIN lsif_last_retention_scan lrs ON lrs.repository_id = cr.id
	JOIN lsif_dirty_repositories dr ON dr.repository_id = cr.id

	-- Ignore records that have been checked recently. Note this condition is
	-- true for a null last_retention_scan_at (which has never been checked).
	WHERE (%s - lrs.last_retention_scan_at > (%s * '1 second'::interval)) IS DISTINCT FROM FALSE
	AND dr.update_token = dr.dirty_token
	ORDER BY
		lrs.last_retention_scan_at NULLS FIRST,
		cr.id -- tie breaker
	LIMIT %s
)
INSERT INTO lsif_last_retention_scan (repository_id, last_retention_scan_at)
SELECT r.id, %s::timestamp FROM repositories r
ON CONFLICT (repository_id) DO UPDATE
SET last_retention_scan_at = %s
RETURNING repository_id
`

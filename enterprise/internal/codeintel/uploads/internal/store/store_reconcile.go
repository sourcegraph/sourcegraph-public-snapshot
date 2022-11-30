package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func (s *store) ReconcileCandidates(ctx context.Context, batchSize int) (_ []int, err error) {
	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(reconcileQuery, batchSize)))
}

const reconcileQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'completed'
	ORDER BY u.last_reconcile_at DESC NULLS FIRST, u.id
	LIMIT %s
),
locked_candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE id = ANY(SELECT id FROM candidates)
	ORDER BY u.id
	FOR UPDATE
)
UPDATE lsif_uploads
SET last_reconcile_at = NOW()
WHERE id = ANY(SELECT id FROM locked_candidates)
RETURNING id
`

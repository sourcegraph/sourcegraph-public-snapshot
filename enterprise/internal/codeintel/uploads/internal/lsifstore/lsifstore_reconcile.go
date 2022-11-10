package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func (s *store) IDsWithMeta(ctx context.Context, ids []int) (_ []int, error error) {
	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(idsWithMetaQuery, pq.Array(ids))))
}

const idsWithMetaQuery = `
SELECT m.dump_id FROM lsif_data_metadata m WHERE m.dump_id = ANY(%s)
`

func (s *store) ReconcileCandidates(ctx context.Context, batchSize int) (_ []int, err error) {
	return basestore.ScanInts(s.db.Query(ctx, sqlf.Sprintf(reconcileQuery, batchSize)))
}

const reconcileQuery = `
WITH candidates AS (
	SELECT m.dump_id
	FROM lsif_data_metadata m
	LEFT JOIN codeintel_last_reconcile lr ON lr.dump_id = m.dump_id
	ORDER BY lr.last_reconcile_at DESC NULLS FIRST, m.dump_id
	LIMIT %s
)
INSERT INTO codeintel_last_reconcile
SELECT dump_id, NOW() FROM candidates
ON CONFLICT (dump_id) DO UPDATE
SET last_reconcile_at = NOW()
RETURNING dump_id
`

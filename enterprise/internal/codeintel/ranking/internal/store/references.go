package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *store) InsertReferencesForRanking(
	ctx context.Context,
	rankingGraphKey string,
	batchSize int,
	uploadID int,
	references chan string,
) (err error) {
	ctx, _, endObservation := s.operations.insertReferencesForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.withTransaction(ctx, func(tx *store) error {
		inserter := func(inserter *batch.Inserter) error {
			for symbols := range batchChannel(references, batchSize) {
				if err := inserter.Insert(ctx, uploadID, pq.Array(symbols), rankingGraphKey); err != nil {
					return err
				}
			}

			return nil
		}

		if err := batch.WithInserter(
			ctx,
			tx.db.Handle(),
			"codeintel_ranking_references",
			batch.MaxNumPostgresParameters,
			[]string{"upload_id", "symbol_names", "graph_key"},
			inserter,
		); err != nil {
			return err
		}

		return nil
	})
}

func (s *store) VacuumAbandonedReferences(ctx context.Context, graphKey string, batchSize int) (_ int, err error) {
	ctx, _, endObservation := s.operations.vacuumAbandonedReferences.With(ctx, &err, observation.Args{LogFields: []otlog.Field{}})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(vacuumAbandonedReferencesQuery, graphKey, graphKey, batchSize)))
	return count, err
}

const vacuumAbandonedReferencesQuery = `
WITH
locked_references AS (
	SELECT id
	FROM codeintel_ranking_references
	WHERE (graph_key < %s OR graph_key > %s)
	ORDER BY graph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_references AS (
	DELETE FROM codeintel_ranking_references
	WHERE id IN (SELECT id FROM locked_references)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_references
`

func (s *store) SoftDeleteStaleReferences(ctx context.Context, graphKey string) (
	numReferenceRecordsScanned int,
	numStaleReferenceRecordsDeleted int,
	err error,
) {
	ctx, _, endObservation := s.operations.softDeleteStaleReferences.With(ctx, &err, observation.Args{LogFields: []otlog.Field{}})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		softDeleteStaleReferencesQuery,
		graphKey, int(threshold/time.Hour), vacuumBatchSize,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(
			&numReferenceRecordsScanned,
			&numStaleReferenceRecordsDeleted,
		); err != nil {
			return 0, 0, err
		}
	}

	return numReferenceRecordsScanned, numStaleReferenceRecordsDeleted, nil
}

const softDeleteStaleReferencesQuery = `
WITH
locked_references AS (
	SELECT
		rr.id,
		rr.upload_id
	FROM codeintel_ranking_references rr
	WHERE
		rr.graph_key = %s AND
		rr.deleted_at IS NULL AND
		(rr.last_scanned_at IS NULL OR NOW() - rr.last_scanned_at >= %s * '1 hour'::interval)
	ORDER BY rr.last_scanned_at ASC NULLS FIRST, rr.id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
candidates AS (
	SELECT
		lr.id,
		uvt.is_default_branch IS TRUE AS safe
	FROM locked_references lr
	LEFT JOIN lsif_uploads u ON u.id = lr.upload_id
	LEFT JOIN lsif_uploads_visible_at_tip uvt ON uvt.repository_id = u.repository_id AND uvt.upload_id = lr.upload_id
),
updated_references AS (
	UPDATE codeintel_ranking_references
	SET last_scanned_at = NOW()
	WHERE id IN (SELECT c.id FROM candidates c WHERE c.safe)
),
deleted_references AS (
	UPDATE codeintel_ranking_references
	SET deleted_at = NOW()
	WHERE id IN (SELECT c.id FROM candidates c WHERE NOT c.safe)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM candidates),
	(SELECT COUNT(*) FROM deleted_references)
`

func (s *store) VacuumDeletedReferences(ctx context.Context, derivativeGraphKey string) (
	numReferenceRecordsDeleted int,
	err error,
) {
	ctx, _, endObservation := s.operations.vacuumDeletedReferences.With(ctx, &err, observation.Args{LogFields: []otlog.Field{}})
	defer endObservation(1, observation.Args{})

	graphKey, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		vacuumDeletedReferencesQuery,
		graphKey,
		derivativeGraphKey,
		vacuumBatchSize,
	)))
	return count, err
}

const vacuumDeletedReferencesQuery = `
WITH
locked_references AS (
	SELECT rr.id
	FROM codeintel_ranking_references rr
	WHERE
		rr.graph_key = %s AND
		rr.deleted_at IS NOT NULL AND
		NOT EXISTS (
			SELECT 1
			FROM codeintel_ranking_progress crp
			WHERE
				crp.graph_key = %s AND
				crp.mapper_completed_at IS NULL
		)
	ORDER BY rr.id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_references AS (
	DELETE FROM codeintel_ranking_references
	WHERE id IN (SELECT id FROM locked_references)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_references
`

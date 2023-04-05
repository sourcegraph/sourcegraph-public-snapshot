package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

func (s *store) VacuumStaleReferences(ctx context.Context, graphKey string) (
	numReferenceRecordsScanned int,
	numStaleReferenceRecordsDeleted int,
	err error,
) {
	ctx, _, endObservation := s.operations.vacuumStaleReferences.With(ctx, &err, observation.Args{LogFields: []otlog.Field{}})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		vacuumStaleReferencesQuery,
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

const vacuumStaleReferencesQuery = `
WITH
locked_references AS (
	SELECT
		rr.id,
		u.repository_id,
		rr.upload_id
	FROM codeintel_ranking_references rr
	JOIN lsif_uploads u ON u.id = rr.upload_id
	WHERE
		rr.graph_key = %s AND
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
	LEFT JOIN lsif_uploads_visible_at_tip uvt ON uvt.repository_id = lr.repository_id AND uvt.upload_id = lr.upload_id
),
updated_references AS (
	UPDATE codeintel_ranking_references
	SET last_scanned_at = NOW()
	WHERE id IN (SELECT c.id FROM candidates c WHERE c.safe)
),
deleted_references AS (
	DELETE FROM codeintel_ranking_references
	WHERE id IN (SELECT c.id FROM candidates c WHERE NOT c.safe)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM candidates),
	(SELECT COUNT(*) FROM deleted_references)
`

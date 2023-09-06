package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *store) GetUploadsForRanking(ctx context.Context, graphKey, objectPrefix string, batchSize int) (_ []shared.ExportedUpload, err error) {
	ctx, _, endObservation := s.operations.getUploadsForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return scanUploads(s.db.Query(ctx, sqlf.Sprintf(
		getUploadsForRankingQuery,
		graphKey,
		batchSize,
		graphKey,
	)))
}

const getUploadsForRankingQuery = `
WITH candidates AS (
	SELECT
		u.id AS upload_id,
		u.repository_id,
		r.name AS repository_name,
		u.root,
		md5(u.repository_id || ':' || u.root || ':' || u.indexer) AS upload_key
	FROM lsif_uploads u
	JOIN lsif_uploads_visible_at_tip uvt ON uvt.upload_id = u.id
	JOIN repo r ON r.id = u.repository_id
	WHERE
		uvt.is_default_branch AND
		r.deleted_at IS NULL AND
		r.blocked IS NULL AND
		NOT EXISTS (
			SELECT 1
			FROM codeintel_ranking_exports re
			WHERE
				re.graph_key = %s AND
				re.upload_id = u.id
		)
	ORDER BY u.id DESC
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
inserted AS (
	INSERT INTO codeintel_ranking_exports (graph_key, upload_id, upload_key)
	SELECT %s, upload_id, upload_key FROM candidates
	ON CONFLICT (graph_key, upload_id) DO NOTHING
	RETURNING id, upload_id
)
SELECT
	i.upload_id,
	i.id,
	c.repository_name,
	c.repository_id,
	c.root
FROM inserted i
JOIN candidates c ON c.upload_id = i.upload_id
ORDER BY c.upload_id
`

var scanUploads = basestore.NewSliceScanner(func(s dbutil.Scanner) (u shared.ExportedUpload, _ error) {
	err := s.Scan(&u.UploadID, &u.ExportedUploadID, &u.Repo, &u.RepoID, &u.Root)
	return u, err
})

func (s *store) VacuumAbandonedExportedUploads(ctx context.Context, graphKey string, batchSize int) (_ int, err error) {
	ctx, _, endObservation := s.operations.vacuumAbandonedExportedUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(vacuumAbandonedExportedUploadsQuery, graphKey, graphKey, batchSize)))
	return count, err
}

const vacuumAbandonedExportedUploadsQuery = `
WITH
locked_exported_uploads AS (
	SELECT id
	FROM codeintel_ranking_exports
	WHERE (graph_key < %s OR graph_key > %s)
	ORDER BY graph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_uploads AS (
	DELETE FROM codeintel_ranking_exports
	WHERE id IN (SELECT id FROM locked_exported_uploads)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_uploads
`

func (s *store) SoftDeleteStaleExportedUploads(ctx context.Context, graphKey string) (
	numExportedUploadRecordsScanned int,
	numStaleExportedUploadRecordsDeleted int,
	err error,
) {
	ctx, _, endObservation := s.operations.softDeleteStaleExportedUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		softDeleteStaleExportedUploadsQuery,
		graphKey, int(threshold/time.Hour), vacuumBatchSize,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(
			&numExportedUploadRecordsScanned,
			&numStaleExportedUploadRecordsDeleted,
		); err != nil {
			return 0, 0, err
		}
	}

	return numExportedUploadRecordsScanned, numStaleExportedUploadRecordsDeleted, nil
}

const softDeleteStaleExportedUploadsQuery = `
WITH
locked_exported_uploads AS (
	SELECT
		cre.id,
		cre.upload_id
	FROM codeintel_ranking_exports cre
	WHERE
		cre.graph_key = %s AND
		cre.deleted_at IS NULL AND
		(cre.last_scanned_at IS NULL OR NOW() - cre.last_scanned_at >= %s * '1 hour'::interval)
	ORDER BY cre.last_scanned_at ASC NULLS FIRST, cre.id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
candidates AS (
	SELECT
		leu.id,
		uvt.is_default_branch IS TRUE AS safe
	FROM locked_exported_uploads leu
	LEFT JOIN lsif_uploads u ON u.id = leu.upload_id
	LEFT JOIN lsif_uploads_visible_at_tip uvt ON uvt.repository_id = u.repository_id AND uvt.upload_id = leu.upload_id AND uvt.is_default_branch
),
updated_exported_uploads AS (
	UPDATE codeintel_ranking_exports cre
	SET last_scanned_at = NOW()
	WHERE id IN (SELECT c.id FROM candidates c WHERE c.safe)
),
deleted_exported_uploads AS (
	UPDATE codeintel_ranking_exports cre
	SET deleted_at = NOW()
	WHERE id IN (SELECT c.id FROM candidates c WHERE NOT c.safe)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM candidates),
	(SELECT COUNT(*) FROM deleted_exported_uploads)
`

func (s *store) VacuumDeletedExportedUploads(ctx context.Context, derivativeGraphKey string) (
	numExportedUploadRecordsDeleted int,
	err error,
) {
	ctx, _, endObservation := s.operations.vacuumDeletedExportedUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	graphKey, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		vacuumDeletedExportedUploadsQuery,
		graphKey,
		derivativeGraphKey,
		vacuumBatchSize,
	)))
	return count, err
}

const vacuumDeletedExportedUploadsQuery = `
WITH
locked_exported_uploads AS (
	SELECT cre.id
	FROM codeintel_ranking_exports cre
	WHERE
		cre.graph_key = %s AND
		cre.deleted_at IS NOT NULL AND
		NOT EXISTS (
			SELECT 1
			FROM codeintel_ranking_progress crp
			WHERE
				crp.graph_key = %s AND
				crp.reducer_completed_at IS NULL AND
				crp.mappers_started_at <= cre.deleted_at
		)
	ORDER BY cre.id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_exported_uploads AS (
	DELETE FROM codeintel_ranking_exports
	WHERE id IN (SELECT id FROM locked_exported_uploads)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_exported_uploads
`

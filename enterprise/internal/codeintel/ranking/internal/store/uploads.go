package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) GetUploadsForRanking(ctx context.Context, graphKey, objectPrefix string, batchSize int) (_ []shared.ExportedUpload, err error) {
	ctx, _, endObservation := s.operations.getUploadsForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return scanUploads(s.db.Query(ctx, sqlf.Sprintf(
		getUploadsForRankingQuery,
		graphKey,
		batchSize,
		graphKey,
		objectPrefix+"/"+graphKey,
		objectPrefix+"/"+graphKey,
	)))
}

const getUploadsForRankingQuery = `
WITH candidates AS (
	SELECT
		u.id,
		u.repository_id,
		r.name AS repository_name,
		u.root
	FROM lsif_uploads u
	JOIN repo r ON r.id = u.repository_id
	WHERE
		u.id IN (
			SELECT uvt.upload_id
			FROM lsif_uploads_visible_at_tip uvt
			WHERE
				uvt.is_default_branch AND
				NOT EXISTS (
					SELECT 1
					FROM codeintel_ranking_exports re
					WHERE
						re.graph_key = %s AND
						re.upload_id = uvt.upload_id
				)
		) AND
		r.deleted_at IS NULL AND
		r.blocked IS NULL
	ORDER BY u.id DESC
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
inserted AS (
	INSERT INTO codeintel_ranking_exports (upload_id, graph_key, object_prefix)
	SELECT
		id,
		%s,
		%s || '/' || id
	FROM candidates
	ON CONFLICT (upload_id, graph_key) DO NOTHING
	RETURNING upload_id AS id
)
SELECT
	c.id,
	c.repository_name,
	c.repository_id,
	c.root,
	%s || '/' || c.id AS object_prefix
FROM candidates c
WHERE c.id IN (SELECT id FROM inserted)
ORDER BY c.id
`

var scanUploads = basestore.NewSliceScanner(func(s dbutil.Scanner) (u shared.ExportedUpload, _ error) {
	err := s.Scan(&u.ID, &u.Repo, &u.RepoID, &u.Root, &u.ObjectPrefix)
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

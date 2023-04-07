package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

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

func (s *store) ProcessStaleExportedUploads(
	ctx context.Context,
	graphKey string,
	batchSize int,
	deleter func(ctx context.Context, objectPrefix string) error,
) (totalDeleted int, err error) {
	ctx, _, endObservation := s.operations.processStaleExportedUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var a int
	err = s.withTransaction(ctx, func(tx *store) error {
		prefixByIDs, err := scanIntStringMap(tx.db.Query(ctx, sqlf.Sprintf(selectStaleExportedUploadsQuery, graphKey, batchSize)))
		if err != nil {
			return err
		}

		ids := make([]int, 0, len(prefixByIDs))
		for id, prefix := range prefixByIDs {
			if err := deleter(ctx, prefix); err != nil {
				return err
			}

			ids = append(ids, id)
		}

		if err := tx.db.Exec(ctx, sqlf.Sprintf(deleteStaleExportedUploadsQuery, pq.Array(ids))); err != nil {
			return err
		}

		a = len(ids)
		return nil
	})
	return a, err
}

var scanIntStringMap = basestore.NewMapScanner(func(s dbutil.Scanner) (k int, v string, _ error) {
	err := s.Scan(&k, &v)
	return k, v, err
})

const selectStaleExportedUploadsQuery = `
SELECT
	re.id,
	re.object_prefix
FROM codeintel_ranking_exports re
WHERE
	re.graph_key = %s AND (re.upload_id IS NULL OR re.upload_id NOT IN (
		SELECT uvt.upload_id
		FROM lsif_uploads_visible_at_tip uvt
		WHERE uvt.is_default_branch
	))
ORDER BY re.upload_id DESC
LIMIT %s
FOR UPDATE OF re SKIP LOCKED
`

const deleteStaleExportedUploadsQuery = `
DELETE FROM codeintel_ranking_exports re
WHERE re.id = ANY(%s)
`

package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type ExportedUpload struct {
	ID           int
	Repo         string
	Root         string
	ObjectPrefix string
}

var scanUploads = basestore.NewSliceScanner(func(s dbutil.Scanner) (u ExportedUpload, _ error) {
	err := s.Scan(&u.ID, &u.Repo, &u.Root, &u.ObjectPrefix)
	return u, err
})

func (s *store) GetUploadsForRanking(ctx context.Context, graphKey, objectPrefix string, batchSize int) (_ []ExportedUpload, err error) {
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
	c.root,
	%s || '/' || c.id AS object_prefix
FROM candidates c
WHERE c.id IN (SELECT id FROM inserted)
ORDER BY c.id
`

func (s *store) ProcessStaleExportedUploads(
	ctx context.Context,
	graphKey string,
	batchSize int,
	deleter func(ctx context.Context, objectPrefix string) error,
) (totalDeleted int, err error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	prefixByIDs, err := scanIntStringMap(tx.Query(ctx, sqlf.Sprintf(selectStaleExportedUploadsQuery, graphKey, batchSize)))
	if err != nil {
		return 0, err
	}

	ids := make([]int, 0, len(prefixByIDs))
	for id, prefix := range prefixByIDs {
		if err := deleter(ctx, prefix); err != nil {
			return 0, err
		}

		ids = append(ids, id)
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(deleteStaleExportedUploadsQuery, pq.Array(ids))); err != nil {
		return 0, err
	}

	return len(ids), nil
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

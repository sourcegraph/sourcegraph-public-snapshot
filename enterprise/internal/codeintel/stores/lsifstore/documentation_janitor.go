package lsifstore

import (
	"context"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func (s *Store) DeleteOldPublicSearchRecords(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int) (int, error) {
	return s.deleteOldSearchRecords(ctx, minimumTimeSinceLastCheck, limit, "public", timeutil.Now())
}

func (s *Store) DeleteOldPrivateSearchRecords(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int) (int, error) {
	return s.deleteOldSearchRecords(ctx, minimumTimeSinceLastCheck, limit, "private", timeutil.Now())
}

func (s *Store) deleteOldSearchRecords(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, tableSuffix string, now time.Time) (_ int, err error) {
	ctx, endObservation := s.operations.deleteOldSearchRecords.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	now = now.UTC()
	interval := int(minimumTimeSinceLastCheck / time.Second)

	numRecords, _, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(deleteOldSearchRecordsQuery, "$SUFFIX", tableSuffix),
		now,
		interval,
		limit,
		now,
	)))
	if err != nil {
		return 0, err
	}

	return numRecords, nil
}

const deleteOldSearchRecordsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/documentation_janitor.go:deleteOldSearchRecords
WITH
candidate_batch AS (
	SELECT id, repo_id, dump_root, lang_name_id, created_at, dump_id
	FROM lsif_data_docs_search_current_$SUFFIX
	WHERE (%s - last_cleanup_scan_at > (%s * '1 second'::interval))
	ORDER BY last_cleanup_scan_at
	LIMIT %s
),
locked_candidate_batch AS (
	SELECT id FROM candidate_batch
	-- TODO - document
	ORDER BY id FOR UPDATE
),
candidates AS (
	SELECT id, repo_id, dump_root, lang_name_id, dump_id
	FROM candidate_batch cb
	WHERE created_at != (
		SELECT MAX(created_at)
		FROM lsif_data_docs_search_current_$SUFFIX cp
		WHERE
			cp.repo_id = cb.repo_id AND
			cp.dump_root = cb.dump_root AND
			cp.lang_name_id = cb.lang_name_id
	)
	ORDER BY id
),
locked_search_records AS (
	SELECT id
	FROM lsif_data_docs_search_$SUFFIX
	WHERE
		(repo_id, dump_root, lang_name_Id, dump_id) IN (
			SELECT repo_id, dump_root, lang_name_Id, dump_id
			FROM candidates
		)
	-- TODO - document
	ORDER BY id FOR UPDATE
),
delete_current_markers AS (
	-- TODO - document
	DELETE FROM lsif_data_docs_search_current_$SUFFIX
	WHERE id IN (SELECT id FROM locked_candidate_batch)
	RETURNING 1
),
delete_search_records AS (
	-- TODO - document
	DELETE FROM lsif_data_docs_search_$SUFFIX
	WHERE id IN (SELECT id FROM locked_search_records)
	RETURNING 1
),
update_timestamp AS (
	-- TODO - document
	UPDATE lsif_data_docs_search_current_$SUFFIX
	SET last_cleanup_scan_at = %s
	-- TODO - document
	WHERE id IN (SELECT id FROM locked_candidate_batch EXCEPT SELECT id FROM candidates)
)
SELECT COUNT(*) FROM delete_search_records
`

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
	ctx, _, endObservation := s.operations.deleteOldSearchRecords.With(ctx, &err, observation.Args{})
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
	-- Select the next batch of records to process (may be all clean records). We do this by order
	-- of last_cleanup_scan_at first, then apply the batch size limit, then lock the records in the
	-- CTE below in a stable (non-temporal) order.

	SELECT id, repo_id, dump_root, lang_name_id, created_at, dump_id
	FROM lsif_data_docs_search_current_$SUFFIX
	WHERE (%s - last_cleanup_scan_at > (%s * '1 second'::interval))
	ORDER BY last_cleanup_scan_at
	LIMIT %s
),
locked_candidate_batch AS (
	SELECT id FROM candidate_batch

	-- Lock the rows underlying the target CTE in a deterministic order so that we don't deadlock
	-- with other processes updating the lsif_data_docs_search_current_* tables. Note that we are
	-- the only consumer of that CTE, so it should be executed inline (and not materialized).
	ORDER BY id FOR UPDATE
),
candidates AS (
	-- Trim down the set of candidates that are NOT the most recently created rows for the same set
	-- of (repository identifier, dump root, language) values. This indicates the set of search items
	-- that have since been replaced and can be removed.

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
	WHERE (repo_id, dump_root, lang_name_Id, dump_id) IN (
		SELECT repo_id, dump_root, lang_name_Id, dump_id
		FROM candidates
	)

	-- Find the rows we'll be deleting from the associated search records table, lock them
	-- in a deterministic order so that we don't deadlock with other processes updating the
	-- lsif_data_docs_search_* tables.
	ORDER BY id FOR UPDATE
),
delete_search_records AS (
	-- Delete search records
	DELETE FROM lsif_data_docs_search_$SUFFIX
	WHERE id IN (SELECT id FROM locked_search_records)
	RETURNING 1
),
delete_current_markers AS (
	-- Delete search marker
	DELETE FROM lsif_data_docs_search_current_$SUFFIX
	WHERE id IN (SELECT id FROM locked_candidate_batch)
	RETURNING 1
),
update_timestamp AS (
	-- Update the last scanned values for the records in the batch
	UPDATE lsif_data_docs_search_current_$SUFFIX
	SET last_cleanup_scan_at = %s

	-- We need to exclude the set of ids we deleted in the sibling CTE delete_current_markers.
	-- If we remove this EXCEPT clause then we just re-insert a new record with identical data
	-- as the one we are trying to delete.
	WHERE id IN (SELECT id FROM locked_candidate_batch EXCEPT SELECT id FROM candidates)
)
-- Return count of search records deleted
SELECT COUNT(*) FROM delete_search_records
`

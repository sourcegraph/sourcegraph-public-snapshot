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
candidates AS (
	SELECT repo_id, dump_root, lang_name_id, dump_id
	FROM lsif_data_docs_search_current_$SUFFIX
	WHERE (%s - last_cleanup_scan_at > (%s * '1 second'::interval))
	ORDER BY repo_id, dump_root, lang_name_id FOR UPDATE
	LIMIT %s
),
deletion_candidates AS (
	SELECT id
	FROM lsif_data_docs_search_$SUFFIX s
	JOIN candidates c
	ON
		c.repo_id = s.repo_id AND
		c.dump_root = s.dump_root AND
		c.lang_name_id = s.lang_name_id
	WHERE
		s.dump_id != c.dump_id
	ORDER BY id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_data_docs_search_$SUFFIX
	WHERE id IN (SELECT id FRom deletion_candidates)
	RETURNING 1
),
update AS (
	UPDATE lsif_data_docs_search_current_$SUFFIX
	SET last_cleanup_scan_at = %s
	WHERE (repo_id, dump_root, lang_name_id) IN (SELECT repo_id, dump_root, lang_name_id FROM candidates)
)
SELECT COUNT(*) FROM deleted
`

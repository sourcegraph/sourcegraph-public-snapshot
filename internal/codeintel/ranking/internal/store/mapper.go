package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *store) InsertPathCountInputs(
	ctx context.Context,
	derivativeGraphKey string,
	batchSize int,
) (
	numReferenceRecordsProcessed int,
	numInputsInserted int,
	err error,
) {
	ctx, _, endObservation := s.operations.insertPathCountInputs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	graphKey, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		insertPathCountInputsQuery,
		derivativeGraphKey,
		graphKey,
		derivativeGraphKey,
		batchSize,
		derivativeGraphKey,
		graphKey,
		graphKey,
		derivativeGraphKey,
		graphKey,
		graphKey,
		derivativeGraphKey,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(
			&numReferenceRecordsProcessed,
			&numInputsInserted,
		); err != nil {
			return 0, 0, err
		}
	}

	return numReferenceRecordsProcessed, numInputsInserted, nil
}

const insertPathCountInputsQuery = `
WITH
progress AS (
	SELECT
		crp.id,
		crp.max_export_id,
		crp.reference_cursor_export_deleted_at,
		crp.reference_cursor_export_id,
		crp.mappers_started_at as started_at
	FROM codeintel_ranking_progress crp
	WHERE
		crp.graph_key = %s AND
		crp.mapper_completed_at IS NULL
),
exported_uploads AS (
	SELECT
		cre.id,
		cre.upload_id,
		cre.upload_key,
		cre.deleted_at
	FROM codeintel_ranking_exports cre
	JOIN progress p ON TRUE
	WHERE
		cre.graph_key = %s AND

		-- Note that we do a check in the processable_symbols CTE below that will
		-- ensure that we don't process a record AND the one it shadows. We end up
		-- taking the lowest ID and no-oping any others that happened to fall into
		-- the window.

		-- Ensure that the record is within the bounds where it would be visible
		-- to the current "snapshot" defined by the ranking computation state row.
		cre.id <= p.max_export_id AND
		(cre.deleted_at IS NULL OR cre.deleted_at > p.started_at) AND

		-- Perf improvement: filter out any uploads that have already been completely
		-- processed. We order uploads by (deleted_at DESC NULLS FIRST, id) as we scan
		-- for candidates. We track the last values we see in each batch so that we can
		-- efficiently discard candidates we don't need to filter out below.

		-- We've already processed all non-deleted exports
		NOT (p.reference_cursor_export_deleted_at IS NOT NULL AND cre.deleted_at IS NULL) AND
		-- We've already processed exports deleted after this point
		NOT (p.reference_cursor_export_deleted_at IS NOT NULL AND cre.deleted_at IS NOT NULL AND p.reference_cursor_export_deleted_at < cre.deleted_at) AND
		NOT (
			p.reference_cursor_export_id IS NOT NULL AND
			-- For records with this deleted_at timestamp (also captures NULL <> NULL match)
			p.reference_cursor_export_deleted_at IS NOT DISTINCT FROM cre.deleted_at AND
			-- Already processed this exported upload
			cre.id < p.reference_cursor_export_id
		)
	ORDER BY cre.graph_key, cre.deleted_at DESC NULLS FIRST, cre.id
),
refs AS (
	SELECT
		rr.id,
		eu.upload_id,
		eu.deleted_at AS exported_upload_deleted_at,
		eu.id AS exported_upload_id,
		eu.upload_key,
		rr.symbol_checksums
	FROM codeintel_ranking_references rr
	JOIN exported_uploads eu ON eu.id = rr.exported_upload_id
	WHERE
		-- Ensure the record isn't already processed
		NOT EXISTS (
			SELECT 1
			FROM codeintel_ranking_references_processed rrp
			WHERE
				rrp.graph_key = %s AND
				rrp.codeintel_ranking_reference_id = rr.id
		)
	ORDER BY eu.deleted_at DESC NULLS FIRST, eu.id, rr.exported_upload_id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
ordered_refs AS (
	SELECT
		r.*,
		-- Rank opposite of the sort order used in the refs CTE above
		RANK() OVER (ORDER BY r.exported_upload_deleted_at ASC NULLS LAST, r.exported_upload_id DESC) AS rank
	FROM refs r
),
locked_refs AS (
	INSERT INTO codeintel_ranking_references_processed (graph_key, codeintel_ranking_reference_id)
	SELECT %s, r.id FROM refs r
	ON CONFLICT DO NOTHING
	RETURNING codeintel_ranking_reference_id
),
referenced_upload_keys AS (
	SELECT DISTINCT r.upload_key
	FROM locked_refs lr
	JOIN refs r ON r.id = lr.codeintel_ranking_reference_id
),
processed_upload_keys AS (
	SELECT cre2.upload_key, cre2.upload_id
	FROM codeintel_ranking_exports cre2
	JOIN codeintel_ranking_references rr2 ON rr2.exported_upload_id = cre2.id
	JOIN codeintel_ranking_references_processed rrp2 ON rrp2.codeintel_ranking_reference_id = rr2.id
	WHERE
		cre2.graph_key = %s AND
		rr2.graph_key = %s AND
		rrp2.graph_key = %s AND
		cre2.upload_key IN (SELECT upload_key FROM referenced_upload_keys)
),
processable_symbols AS (
	SELECT r.symbol_checksums
	FROM locked_refs lr
	JOIN refs r ON r.id = lr.codeintel_ranking_reference_id
	WHERE
		-- Do not re-process references for repository/root/indexers that have already been
		-- processed. We'll still insert a processed reference so that we know we've done the
		-- "work", but we'll simply no-op the counts for this input.
		NOT EXISTS (
			SELECT 1
			FROM processed_upload_keys puk
			WHERE
				puk.upload_key = r.upload_key AND
				puk.upload_id != r.upload_id
		) AND

		-- For multiple references for the same repository/root/indexer in THIS batch, we want to
		-- process the one associated with the most recently processed upload record. This should
		-- maximize fresh results.
		NOT EXISTS (
			SELECT 1
			FROM locked_refs lr2
			JOIN refs r2 ON r2.id = lr2.codeintel_ranking_reference_id
			WHERE
				r2.upload_key = r.upload_key AND
				r2.upload_id > r.upload_id
		)
),
referenced_symbols AS (
	SELECT DISTINCT unnest(r.symbol_checksums) AS symbol_checksum
	FROM processable_symbols r
),
ranked_referenced_definitions AS (
	SELECT
		rd.id AS definition_id,

		-- Group by repository/root/indexer and order by descending ids. We
		-- will only count the rows with rank = 1 in the outer query in order
		-- to break ties when shadowed definitions are present.
		RANK() OVER (PARTITION BY cre.upload_key ORDER BY cre.upload_id DESC) AS rank
	FROM codeintel_ranking_definitions rd
	JOIN referenced_symbols rs ON rs.symbol_checksum = rd.symbol_checksum
	JOIN codeintel_ranking_exports cre ON cre.id = rd.exported_upload_id
	JOIN progress p ON TRUE
	WHERE
		rd.graph_key = %s AND
		cre.graph_key = %s AND

		-- Note that we do a check in the processable_symbols CTE below that will
		-- ensure that we don't process a record AND the one it shadows. We end up
		-- taking the lowest ID and no-oping any others that happened to fall into
		-- the window.

		-- Ensure that the record is within the bounds where it would be visible
		-- to the current "snapshot" defined by the ranking computation state row.
		cre.id <= p.max_export_id AND
		(cre.deleted_at IS NULL OR cre.deleted_at > p.started_at)
	ORDER BY cre.graph_key, cre.deleted_at DESC NULLS FIRST, cre.id
),
referenced_definitions AS (
	SELECT
		s.definition_id,
		COUNT(*) AS count
	FROM ranked_referenced_definitions s

	-- For multiple uploads in the same repository/root/indexer, only consider
	-- definition records attached to the one with the highest id. This should
	-- prevent over-counting definitions when there are multiple uploads in the
	-- exported set, but the shadowed (newly non-visible) uploads have not yet
	-- been removed by the janitor processes.
	WHERE s.rank = 1
	GROUP BY s.definition_id
),
ins AS (
	INSERT INTO codeintel_ranking_path_counts_inputs AS target (graph_key, definition_id, count, processed)
	SELECT
		%s,
		rx.definition_id,
		rx.count,
		false
	FROM referenced_definitions rx
	ON CONFLICT (graph_key, definition_id) WHERE NOT processed DO UPDATE SET count = target.count + EXCLUDED.count
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_ranking_progress
	SET
		-- Update cursor values with the last item in the batch
		reference_cursor_export_deleted_at = COALESCE((SELECT tor.exported_upload_deleted_at FROM ordered_refs tor WHERE tor.rank = 1 LIMIT 1), NULL),
		reference_cursor_export_id         = COALESCE((SELECT tor.exported_upload_id FROM ordered_refs tor WHERE tor.rank = 1 LIMIT 1), NULL),
		-- Update overall progress
		num_reference_records_processed    = COALESCE(num_reference_records_processed, 0) + (SELECT COUNT(*) FROM locked_refs),
		mapper_completed_at                = CASE WHEN (SELECT COUNT(*) FROM refs) = 0 THEN NOW() ELSE NULL END

	WHERE id IN (SELECT id FROM progress)
)
SELECT
	(SELECT COUNT(*) FROM locked_refs),
	(SELECT COUNT(*) FROM ins)
`

func (s *store) InsertInitialPathCounts(
	ctx context.Context,
	derivativeGraphKey string,
	batchSize int,
) (
	numInitialPathsProcessed int,
	numInitialPathRanksInserted int,
	err error,
) {
	ctx, _, endObservation := s.operations.insertInitialPathCounts.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	graphKey, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		insertInitialPathCountsInputsQuery,
		derivativeGraphKey,
		graphKey,
		derivativeGraphKey,
		batchSize,
		derivativeGraphKey,
		derivativeGraphKey,
		graphKey,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(
			&numInitialPathsProcessed,
			&numInitialPathRanksInserted,
		); err != nil {
			return 0, 0, err
		}
	}

	return numInitialPathsProcessed, numInitialPathRanksInserted, nil
}

const insertInitialPathCountsInputsQuery = `
WITH
progress AS (
	SELECT
		crp.id,
		crp.max_export_id,
		crp.path_cursor_deleted_export_at,
		crp.path_cursor_export_id,
		crp.mappers_started_at as started_at
	FROM codeintel_ranking_progress crp
	WHERE
		crp.graph_key = %s AND
		crp.seed_mapper_completed_at IS NULL
),
exported_uploads AS (
	SELECT
		cre.id,
		cre.upload_id,
		cre.deleted_at
	FROM codeintel_ranking_exports cre
	JOIN progress p ON TRUE
	WHERE
		cre.graph_key = %s AND

		-- Note that we do a check in the processable_symbols CTE below that will
		-- ensure that we don't process a record AND the one it shadows. We end up
		-- taking the lowest ID and no-oping any others that happened to fall into
		-- the window.

		-- Ensure that the record is within the bounds where it would be visible
		-- to the current "snapshot" defined by the ranking computation state row.
		cre.id <= p.max_export_id AND
		(cre.deleted_at IS NULL OR cre.deleted_at > p.started_at) AND

		-- Perf improvement: filter out any uploads that have already been completely
		-- processed. We order uploads by (deleted_at DESC NULLS FIRST, id) as we scan
		-- for candidates. We track the last values we see in each batch so that we can
		-- efficiently discard candidates we don't need to filter out below.

		-- We've already processed all non-deleted exports
		NOT (p.path_cursor_deleted_export_at IS NOT NULL AND cre.deleted_at IS NULL) AND
		-- We've already processed exports deleted after this point
		NOT (p.path_cursor_deleted_export_at IS NOT NULL AND cre.deleted_at IS NOT NULL AND p.path_cursor_deleted_export_at < cre.deleted_at) AND
		NOT (
			p.path_cursor_export_id IS NOT NULL AND
			-- For records with this deleted_at timestamp (also captures NULL <> NULL match)
			p.path_cursor_deleted_export_at IS NOT DISTINCT FROM cre.deleted_at AND
			-- Already processed this exported upload
			cre.id < p.path_cursor_export_id
		)
	ORDER BY cre.graph_key, cre.deleted_at DESC NULLS FIRST, cre.id
),
unprocessed_path_counts AS (
	SELECT
		ipr.id,
		eu.upload_id,
		eu.deleted_at AS exported_upload_deleted_at,
		eu.id AS exported_upload_id,
		ipr.graph_key,
		CASE
			WHEN ipr.document_path != '' THEN array_append('{}'::text[], ipr.document_path)
			ELSE ipr.document_paths
		END AS document_paths
	FROM codeintel_initial_path_ranks ipr
	JOIN exported_uploads eu ON eu.id = ipr.exported_upload_id
	WHERE
		-- Ensure the record isn't already processed
		NOT EXISTS (
			SELECT 1
			FROM codeintel_initial_path_ranks_processed prp
			WHERE
				prp.graph_key = %s AND
				prp.codeintel_initial_path_ranks_id = ipr.id
		)
	ORDER BY eu.deleted_at DESC NULLS FIRST, eu.id, ipr.exported_upload_id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
ordered_paths AS (
	SELECT
		p.*,
		-- Rank opposite of the sort order used in the unprocessed_path_counts CTE above
		RANK() OVER (ORDER BY p.exported_upload_deleted_at ASC NULLS LAST, p.exported_upload_id DESC) AS rank
	FROM unprocessed_path_counts p
),
locked_path_counts AS (
	INSERT INTO codeintel_initial_path_ranks_processed (graph_key, codeintel_initial_path_ranks_id)
	SELECT
		%s,
		eupc.id
	FROM unprocessed_path_counts eupc
	ON CONFLICT DO NOTHING
	RETURNING codeintel_initial_path_ranks_id
),
expanded_unprocessed_path_counts AS (
	SELECT
		upc.id,
		upc.upload_id,
		upc.exported_upload_id,
		upc.graph_key,
		unnest(upc.document_paths) AS document_path
	FROM unprocessed_path_counts upc
),
ins AS (
	INSERT INTO codeintel_ranking_path_counts_inputs (graph_key, definition_id, count, processed)
	SELECT
		%s,
		rd.id,
		0,
		false
	FROM locked_path_counts lpc
	JOIN expanded_unprocessed_path_counts eupc ON eupc.id = lpc.codeintel_initial_path_ranks_id
	JOIN codeintel_ranking_definitions rd ON
		rd.exported_upload_id = eupc.exported_upload_id AND
		rd.document_path = eupc.document_path
	WHERE
		rd.graph_key = %s AND
		-- See definition of sentinelPathDefinitionName
		rd.symbol_checksum = '\xc3e97dd6e97fb5125688c97f36720cbe'::bytea
	ON CONFLICT DO NOTHING
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_ranking_progress
	SET
		-- Update cursor values with the last item in the batch
		path_cursor_deleted_export_at = COALESCE((SELECT op.exported_upload_deleted_at FROM ordered_paths op WHERE op.rank = 1 LIMIT 1), NULL),
		path_cursor_export_id         = COALESCE((SELECT op.exported_upload_id FROM ordered_paths op WHERE op.rank = 1 LIMIT 1), NULL),
		-- Update overall progress
		num_path_records_processed    = COALESCE(num_path_records_processed, 0) + (SELECT COUNT(*) FROM locked_path_counts),
		seed_mapper_completed_at      = CASE WHEN (SELECT COUNT(*) FROM unprocessed_path_counts) = 0 THEN NOW() ELSE NULL END
	WHERE id IN (SELECT id FROM progress)
)
SELECT
	(SELECT COUNT(*) FROM locked_path_counts),
	(SELECT COUNT(*) FROM ins)
`

func (s *store) VacuumStaleProcessedReferences(ctx context.Context, derivativeGraphKey string, batchSize int) (_ int, err error) {
	ctx, _, endObservation := s.operations.vacuumStaleProcessedReferences.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(vacuumStaleProcessedReferencesQuery, derivativeGraphKey, derivativeGraphKey, batchSize)))
	return count, err
}

const vacuumStaleProcessedReferencesQuery = `
WITH
locked_references_processed AS (
	SELECT id
	FROM codeintel_ranking_references_processed
	WHERE (graph_key < %s OR graph_key > %s)
	ORDER BY graph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_locked_references_processed AS (
	DELETE FROM codeintel_ranking_references_processed
	WHERE id IN (SELECT id FROM locked_references_processed)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_locked_references_processed
`

func (s *store) VacuumStaleProcessedPaths(ctx context.Context, derivativeGraphKey string, batchSize int) (_ int, err error) {
	ctx, _, endObservation := s.operations.vacuumStaleProcessedPaths.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(vacuumStaleProcessedPathsQuery, derivativeGraphKey, derivativeGraphKey, batchSize)))
	return count, err
}

const vacuumStaleProcessedPathsQuery = `
WITH
locked_paths_processed AS (
	SELECT id
	FROM codeintel_initial_path_ranks_processed
	WHERE (graph_key < %s OR graph_key > %s)
	ORDER BY graph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_locked_paths_processed AS (
	DELETE FROM codeintel_initial_path_ranks_processed
	WHERE id IN (SELECT id FROM locked_paths_processed)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_locked_paths_processed
`

func (s *store) VacuumStaleGraphs(ctx context.Context, derivativeGraphKey string, batchSize int) (_ int, err error) {
	ctx, _, endObservation := s.operations.vacuumStaleGraphs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(vacuumStaleGraphsQuery, derivativeGraphKey, derivativeGraphKey, batchSize)))
	return count, err
}

const vacuumStaleGraphsQuery = `
WITH
locked_path_counts_inputs AS (
	SELECT id
	FROM codeintel_ranking_path_counts_inputs
	WHERE (graph_key < %s OR graph_key > %s)
	ORDER BY graph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_path_counts_inputs AS (
	DELETE FROM codeintel_ranking_path_counts_inputs
	WHERE id IN (SELECT id FROM locked_path_counts_inputs)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_path_counts_inputs
`

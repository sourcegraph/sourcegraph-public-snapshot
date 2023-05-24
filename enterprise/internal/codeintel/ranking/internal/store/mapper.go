package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
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
		derivativeGraphKey,
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
		crp.mappers_started_at as started_at
	FROM codeintel_ranking_progress crp
	WHERE
		crp.graph_key = %s AND
		crp.mapper_completed_at IS NULL
),
exported_uploads AS (
	SELECT
		cre.id,
		cre.upload_id
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
		(cre.deleted_at IS NULL OR cre.deleted_at > p.started_at)
	ORDER BY cre.graph_key, cre.deleted_at DESC NULLS FIRST, cre.id
),
refs AS (
	SELECT
		rr.id,
		eu.upload_id,
		rr.symbol_names
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
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
locked_refs AS (
	INSERT INTO codeintel_ranking_references_processed (graph_key, codeintel_ranking_reference_id)
	SELECT %s, r.id FROM refs r
	ON CONFLICT DO NOTHING
	RETURNING codeintel_ranking_reference_id
),
processable_symbols AS (
	SELECT r.symbol_names
	FROM locked_refs lr
	JOIN refs r ON r.id = lr.codeintel_ranking_reference_id
	JOIN lsif_uploads u ON u.id = r.upload_id
	WHERE
		-- Do not re-process references for repository/root/indexers that have already been
		-- processed. We'll still insert a processed reference so that we know we've done the
		-- "work", but we'll simply no-op the counts for this input.
		NOT EXISTS (
			SELECT 1
			FROM lsif_uploads u2
			JOIN codeintel_ranking_exports cre2 ON cre2.upload_id = u2.id
			JOIN codeintel_ranking_references rr2 ON rr2.exported_upload_id = cre2.id
			JOIN codeintel_ranking_references_processed rrp2 ON rrp2.codeintel_ranking_reference_id = rr2.id
			WHERE
				rrp2.graph_key = %s AND
				u.repository_id = u2.repository_id AND
				u.root = u2.root AND
				u.indexer = u2.indexer AND
				u.id != u2.id
		) AND

		-- For multiple references for the same repository/root/indexer in THIS batch, we want to
		-- process the one associated with the most recently processed upload record. This should
		-- maximize fresh results.
		NOT EXISTS (
			SELECT 1
			FROM locked_refs lr2
			JOIN refs r2 ON r2.id = lr2.codeintel_ranking_reference_id
			JOIN lsif_uploads u2 ON u2.id = r2.upload_id
			WHERE
				u.repository_id = u2.repository_id AND
				u.root = u2.root AND
				u.indexer = u2.indexer AND
				u.finished_at < u2.finished_at
		)
),
referenced_symbols AS (
	SELECT unnest(r.symbol_names) AS symbol_name
	FROM processable_symbols r
),
referenced_definitions AS (
	SELECT
		s.repository_id,
		s.document_path,
		COUNT(*) AS count
	FROM (
		SELECT
			u.repository_id,
			rd.document_path,

			-- Group by repository/root/indexer and order by descending ids. We
			-- will only count the rows with rank = 1 in the outer query in order
			-- to break ties when shadowed definitions are present.
			RANK() OVER (
				PARTITION BY u.repository_id, u.root, u.indexer
				ORDER BY u.id DESC
			) AS rank
		FROM codeintel_ranking_definitions rd
		JOIN referenced_symbols rs ON rs.symbol_name = rd.symbol_name
		JOIN exported_uploads eu ON eu.id = rd.exported_upload_id
		JOIN lsif_uploads u ON u.id = eu.upload_id
	) s

	-- For multiple uploads in the same repository/root/indexer, only consider
	-- definition records attached to the one with the highest id. This should
	-- prevent over-counting definitions when there are multiple uploads in the
	-- exported set, but the shadowed (newly non-visible) uploads have not yet
	-- been removed by the janitor processes.
	WHERE s.rank = 1
	GROUP BY s.repository_id, s.document_path
),
ins AS (
	INSERT INTO codeintel_ranking_path_counts_inputs (repository_id, document_path, count, graph_key)
	SELECT
		rx.repository_id,
		rx.document_path,
		rx.count,
		%s
	FROM referenced_definitions rx
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_ranking_progress
	SET
		num_reference_records_processed = COALESCE(num_reference_records_processed, 0) + (SELECT COUNT(*) FROM locked_refs),
		mapper_completed_at             = CASE WHEN (SELECT COUNT(*) FROM refs) = 0 THEN NOW() ELSE NULL END
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
		crp.mappers_started_at as started_at
	FROM codeintel_ranking_progress crp
	WHERE
		crp.graph_key = %s AND
		crp.seed_mapper_completed_at IS NULL
),
exported_uploads AS (
	SELECT
		cre.id,
		cre.upload_id
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
		(cre.deleted_at IS NULL OR cre.deleted_at > p.started_at)
	ORDER BY cre.graph_key, cre.deleted_at DESC NULLS FIRST, cre.id
),
unprocessed_path_counts AS (
	SELECT
		ipr.id,
		eu.upload_id,
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
	ORDER BY ipr.id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
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
		upc.graph_key,
		unnest(upc.document_paths) AS document_path
	FROM unprocessed_path_counts upc
),
ins AS (
	INSERT INTO codeintel_ranking_path_counts_inputs (repository_id, document_path, count, graph_key)
	SELECT
		u.repository_id,
		eupc.document_path,
		0,
		%s
	FROM locked_path_counts lpc
	JOIN expanded_unprocessed_path_counts eupc on eupc.id = lpc.codeintel_initial_path_ranks_id
	JOIN lsif_uploads u ON u.id = eupc.upload_id
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_ranking_progress
	SET
		num_path_records_processed = COALESCE(num_path_records_processed, 0) + (SELECT COUNT(*) FROM locked_path_counts),
		seed_mapper_completed_at   = CASE WHEN (SELECT COUNT(*) FROM unprocessed_path_counts) = 0 THEN NOW() ELSE NULL END
	WHERE id IN (SELECT id FROM progress)
)
SELECT
	(SELECT COUNT(*) FROM locked_path_counts),
	(SELECT COUNT(*) FROM ins)
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

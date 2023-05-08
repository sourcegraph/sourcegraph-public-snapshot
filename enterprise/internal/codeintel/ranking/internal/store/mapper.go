package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"

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
		graphKey,
		derivativeGraphKey,
		batchSize,
		derivativeGraphKey,
		derivativeGraphKey,
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
refs AS (
	SELECT
		rr.id,
		rr.upload_id,
		rr.symbol_names
	FROM codeintel_ranking_references rr
	WHERE
		rr.graph_key = %s AND
		NOT EXISTS (
			SELECT 1
			FROM codeintel_ranking_references_processed rrp
			WHERE
				rrp.graph_key = %s AND
				rrp.codeintel_ranking_reference_id = rr.id
		)
	ORDER BY rr.id
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
			JOIN codeintel_ranking_references rr ON rr.upload_id = u2.id
			JOIN codeintel_ranking_references_processed rrp ON rrp.codeintel_ranking_reference_id = rr.id
			WHERE
				rrp.graph_key = %s AND
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
		u.repository_id,
		rd.document_path,
		rd.graph_key,
		COUNT(*) AS count
	FROM codeintel_ranking_definitions rd
	JOIN referenced_symbols rs ON rs.symbol_name = rd.symbol_name
	JOIN lsif_uploads u ON u.id = rd.upload_id
	WHERE rd.graph_key = %s
	GROUP BY u.repository_id, rd.document_path, rd.graph_key
),
ins AS (
	INSERT INTO codeintel_ranking_path_counts_inputs (repository_id, document_path, count, graph_key)
	SELECT
		rx.repository_id,
		rx.document_path,
		SUM(rx.count),
		%s
	FROM referenced_definitions rx
	GROUP BY rx.repository_id, rx.document_path
	RETURNING 1
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
unprocessed_path_counts AS (
	SELECT
		ipr.id,
		ipr.upload_id,
		ipr.graph_key,
		CASE
			WHEN ipr.document_path != '' THEN array_append('{}'::text[], ipr.document_path)
			ELSE ipr.document_paths
		END AS document_paths
	FROM codeintel_initial_path_ranks ipr
	WHERE
		ipr.graph_key = %s AND
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
expanded_unprocessed_path_counts AS (
	SELECT
		upc.id,
		upc.upload_id,
		upc.graph_key,
		unnest(upc.document_paths) AS document_path
	FROM unprocessed_path_counts upc
),
locked_path_counts AS (
	INSERT INTO codeintel_initial_path_ranks_processed (graph_key, codeintel_initial_path_ranks_id)
	SELECT
		%s,
		eupc.id
	FROM expanded_unprocessed_path_counts eupc
	ON CONFLICT DO NOTHING
	RETURNING codeintel_initial_path_ranks_id
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
)
SELECT
	(SELECT COUNT(*) FROM locked_path_counts),
	(SELECT COUNT(*) FROM ins)
`

func (s *store) VacuumStaleGraphs(ctx context.Context, derivativeGraphKey string, batchSize int) (_ int, err error) {
	ctx, _, endObservation := s.operations.vacuumStaleGraphs.With(ctx, &err, observation.Args{LogFields: []otlog.Field{}})
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

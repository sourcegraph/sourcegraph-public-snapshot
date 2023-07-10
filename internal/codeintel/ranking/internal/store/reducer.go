package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *store) InsertPathRanks(
	ctx context.Context,
	derivativeGraphKey string,
	batchSize int,
) (numInputsProcessed int, numPathRanksInserted int, err error) {
	ctx, _, endObservation := s.operations.insertPathRanks.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("derivativeGraphKey", derivativeGraphKey),
	}})
	defer endObservation(1, observation.Args{})

	_, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		insertPathRanksQuery,
		derivativeGraphKey,
		derivativeGraphKey,
		batchSize,
		derivativeGraphKey,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return 0, 0, errors.New("no rows from count")
	}

	if err = rows.Scan(&numInputsProcessed, &numPathRanksInserted); err != nil {
		return 0, 0, err
	}

	return numInputsProcessed, numPathRanksInserted, nil
}

const insertPathRanksQuery = `
WITH
progress AS (
	SELECT crp.id
	FROM codeintel_ranking_progress crp
	WHERE
		crp.graph_key = %s and
		crp.reducer_started_at IS NOT NULL AND
		crp.reducer_completed_at IS NULL
),
rank_ids AS (
	SELECT pci.id
	FROM codeintel_ranking_path_counts_inputs pci
	JOIN progress p ON TRUE
	WHERE
		pci.graph_key = %s AND
		NOT pci.processed
	ORDER BY pci.graph_key, pci.definition_id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
input_ranks AS (
	SELECT
		pci.id,
		u.repository_id,
		rd.document_path AS path,
		pci.count
	FROM codeintel_ranking_path_counts_inputs pci
	JOIN codeintel_ranking_definitions rd ON rd.id = pci.definition_id
	JOIN codeintel_ranking_exports eu ON eu.id = rd.exported_upload_id
	JOIN lsif_uploads u ON u.id = eu.upload_id
	JOIN repo r ON r.id = u.repository_id
	JOIN progress p ON TRUE
	WHERE
		pci.id IN (SELECT id FROM rank_ids) AND
		r.deleted_at IS NULL AND
		r.blocked IS NULL
),
processed AS (
	UPDATE codeintel_ranking_path_counts_inputs
	SET processed = true
	WHERE id IN (SELECT ir.id FROM rank_ids ir)
	RETURNING 1
),
inserted AS (
	INSERT INTO codeintel_path_ranks AS pr (graph_key, repository_id, payload)
	SELECT
		%s,
		temp.repository_id,
		jsonb_object_agg(temp.path, temp.count)
	FROM (
		SELECT
			cr.repository_id,
			cr.path,
			SUM(count) AS count
		FROM input_ranks cr
		GROUP BY cr.repository_id, cr.path
	) temp
	GROUP BY temp.repository_id
	ON CONFLICT (graph_key, repository_id) DO UPDATE SET
		payload = (
			SELECT jsonb_object_agg(key, sum) FROM (
				SELECT key, SUM(value::int) AS sum
				FROM
					(
						SELECT * FROM jsonb_each(pr.payload)
						UNION
						SELECT * FROM jsonb_each(EXCLUDED.payload)
					) AS both_payloads
				GROUP BY key
			) AS combined_json
		)
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_ranking_progress
	SET
		num_count_records_processed = COALESCE(num_count_records_processed, 0) + (SELECT COUNT(*) FROM processed),
		reducer_completed_at        = CASE WHEN (SELECT COUNT(*) FROM rank_ids) = 0 THEN NOW() ELSE NULL END
	WHERE id IN (SELECT id FROM progress)
)
SELECT
	(SELECT COUNT(*) FROM processed) AS num_processed,
	(SELECT COUNT(*) FROM inserted) AS num_inserted
`

func (s *store) VacuumStaleRanks(ctx context.Context, derivativeGraphKey string) (rankRecordsDeleted, rankRecordsScanned int, err error) {
	ctx, _, endObservation := s.operations.vacuumStaleRanks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if _, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey); !ok {
		return 0, 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		vacuumStaleRanksQuery,
		derivativeGraphKey,
	))
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&rankRecordsScanned, &rankRecordsDeleted); err != nil {
			return 0, 0, err
		}
	}

	return rankRecordsScanned, rankRecordsDeleted, nil
}

const vacuumStaleRanksQuery = `
WITH
valid_graph_keys AS (
	-- Select current graph key
	SELECT %s AS graph_key
	-- Select previous graph key
	UNION (
		SELECT crp.graph_key
		FROM codeintel_ranking_progress crp
		WHERE crp.reducer_completed_at IS NOT NULL
		ORDER BY crp.reducer_completed_at DESC
		LIMIT 1
	)
),
locked_records AS (
	-- Lock all path rank records that don't have a valid graph key
	SELECT id
	FROM codeintel_path_ranks
	WHERE graph_key NOT IN (SELECT graph_key FROM valid_graph_keys)
	ORDER BY id
	FOR UPDATE
),
deleted_records AS (
	DELETE FROM codeintel_path_ranks
	WHERE id IN (SELECT id FROM locked_records)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM locked_records),
	(SELECT COUNT(*) FROM deleted_records)
`

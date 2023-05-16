package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
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
	SELECT
		crp.id,
		crp.mappers_started_at as started_at
	FROM codeintel_ranking_progress crp
	WHERE
		crp.graph_key = %s and
		crp.reducer_started_at IS NOT NULL AND
		crp.reducer_completed_at IS NULL
),
input_ranks AS (
	SELECT
		pci.id,
		pci.repository_id,
		pci.document_path AS path,
		pci.count
	FROM codeintel_ranking_path_counts_inputs pci
	JOIN progress p ON TRUE
	WHERE
		pci.graph_key = %s AND
		NOT pci.processed AND
		EXISTS (
			SELECT 1 FROM repo r
			WHERE
				r.id = pci.repository_id AND
				r.deleted_at IS NULL AND
				r.blocked IS NULL
		)
	ORDER BY pci.graph_key, pci.repository_id, pci.id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
processed AS (
	UPDATE codeintel_ranking_path_counts_inputs
	SET processed = true
	WHERE id IN (SELECT ir.id FROM input_ranks ir)
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
	SET reducer_completed_at = NOW()
	WHERE
		id IN (SELECT id FROM progress) AND
		NOT EXISTS (SELECT 1 FROM processed)
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

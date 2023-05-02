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

func (s *store) InsertPathRanks(
	ctx context.Context,
	derivativeGraphKey string,
	batchSize int,
) (numInputsProcessed int, numPathRanksInserted int, err error) {
	ctx, _, endObservation := s.operations.insertPathRanks.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("derivativeGraphKey", derivativeGraphKey),
	}})
	defer endObservation(1, observation.Args{})

	_, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		insertPathRanksQuery,
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
input_ranks AS (
	SELECT
		pci.id,
		pci.repository_id,
		pci.document_path AS path,
		pci.count
	FROM codeintel_ranking_path_counts_inputs pci
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
	INSERT INTO codeintel_path_ranks AS pr (repository_id, graph_key, payload)
	SELECT
		temp.repository_id,
		%s,
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
	ON CONFLICT (repository_id) DO UPDATE SET
		graph_key = EXCLUDED.graph_key,
		payload = CASE
			WHEN pr.graph_key != EXCLUDED.graph_key
				THEN EXCLUDED.payload
			ELSE
				(
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
			END
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM processed) AS num_processed,
	(SELECT COUNT(*) FROM inserted) AS num_inserted
`

func (s *store) VacuumStaleRanks(ctx context.Context, derivativeGraphKey string) (rankRecordsDeleted, rankRecordsScanned int, err error) {
	ctx, _, endObservation := s.operations.vacuumStaleRanks.With(ctx, &err, observation.Args{LogFields: []otlog.Field{}})
	defer endObservation(1, observation.Args{})

	graphKey, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		vacuumStaleRanksQuery,
		derivativeGraphKey,
		graphKey,
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
matching_graph_keys AS (
	SELECT DISTINCT graph_key
	FROM codeintel_path_ranks
	-- Implicit delete anything with a different graph key root
	WHERE graph_key != %s AND graph_key LIKE %s || '.%%'
),
valid_graph_keys AS (
	-- Select the current graph key as well as the highest graph key that
	-- shares the same parent graph key. Returning both will help bridge
	-- the gap that happens if we were to flush the entire table at the
	-- start of a new graph reduction.
	--
	-- This may have the effect of returning stale ranking data for a repo
	-- for which we no longer have SCIP data, but only from the previous
	-- graph reduction (and changing the parent graph key will flush all
	-- previous data (see the CTE definition above) if the need arises.
	SELECT %s AS graph_key
	UNION (
		SELECT graph_key
		FROM matching_graph_keys
		ORDER BY reverse(split_part(reverse(graph_key), '-', 1))::int DESC
		LIMIT 1
	)
),
locked_records AS (
	-- Lock all path rank records that don't have a recent graph key
	SELECT repository_id
	FROM codeintel_path_ranks
	WHERE graph_key NOT IN (SELECT graph_key FROM valid_graph_keys)
	ORDER BY repository_id
	FOR UPDATE
),
del AS (
	DELETE FROM codeintel_path_ranks
	WHERE repository_id IN (SELECT repository_id FROM locked_records)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM locked_records),
	(SELECT COUNT(*) FROM del)
`

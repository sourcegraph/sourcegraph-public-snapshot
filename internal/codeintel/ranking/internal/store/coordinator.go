package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// for lazy mocking in tests
var testNow = time.Now

// MaxProgressRecords is the maximum number of progress records we'll track before pruning
// older entries.
const MaxProgressRecords = 10

func (s *store) Coordinate(
	ctx context.Context,
	derivativeGraphKey string,
) (err error) {
	ctx, _, endObservation := s.operations.coordinate.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	graphKey, ok := rankingshared.GraphKeyFromDerivativeGraphKey(derivativeGraphKey)
	if !ok {
		return errors.Newf("unexpected derivative graph key %q", derivativeGraphKey)
	}

	now := testNow()

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(
		coordinateStartMapperQuery,
		graphKey,
		graphKey,
		graphKey,
		derivativeGraphKey,
		now,
		derivativeGraphKey,
	)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(coordinatePruneQuery, derivativeGraphKey, MaxProgressRecords)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(
		coordinateStartReducerQuery,
		derivativeGraphKey,
		now,
		derivativeGraphKey,
	)); err != nil {
		return err
	}

	return nil
}

const coordinateStartMapperQuery = `
WITH
progress AS (
	SELECT
		COALESCE((SELECT MAX(id) FROM codeintel_ranking_exports WHERE graph_key = %s), 0) AS max_export_id
),
processable_paths AS (
	SELECT ipr.id
	FROM codeintel_initial_path_ranks ipr
	JOIN codeintel_ranking_exports cre ON cre.id = ipr.exported_upload_id
	JOIN progress p ON TRUE
	WHERE
		ipr.graph_key = %s AND
		cre.id <= p.max_export_id AND
		cre.deleted_at IS NULL
),
processable_references AS (
	SELECT rr.id
	FROM codeintel_ranking_references rr
	JOIN codeintel_ranking_exports cre ON cre.id = rr.exported_upload_id
	JOIN progress p ON TRUE
	WHERE
		rr.graph_key = %s AND
		cre.id <= p.max_export_id AND
		cre.deleted_at IS NULL
),
values AS (
	SELECT
		%s,
		p.max_export_id,
		%s::timestamp with time zone,
		(SELECT COUNT(*) FROM processable_paths),
		(SELECT COUNT(*) FROM processable_references)
	FROM progress p
	WHERE NOT EXISTS (
		SELECT 1
		FROM codeintel_ranking_progress
		WHERE graph_key = %s
	)
)
INSERT INTO codeintel_ranking_progress(
	graph_key,
	max_export_id,
	mappers_started_at,
	num_path_records_total,
	num_reference_records_total
)
SELECT * FROM values
ON CONFLICT DO NOTHING
`

const coordinatePruneQuery = `
DELETE FROM codeintel_ranking_progress WHERE id IN (
	SELECT id
	FROM codeintel_ranking_progress
	WHERE graph_key != %s
	ORDER BY mappers_started_at DESC
	OFFSET %s
)
`

const coordinateStartReducerQuery = `
WITH
processable_counts AS (
	SELECT pci.id
	FROM codeintel_ranking_path_counts_inputs pci
	WHERE
		pci.graph_key = %s AND
		NOT pci.processed
)
UPDATE codeintel_ranking_progress
SET
	reducer_started_at      = %s,
	num_count_records_total = (SELECT COUNT(*) FROM processable_counts)
WHERE
	graph_key = %s AND
	mapper_completed_at IS NOT NULL AND
	seed_mapper_completed_at IS NOT NULL AND
	reducer_started_at IS NULL
`

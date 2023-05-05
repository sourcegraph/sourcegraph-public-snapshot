package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

	if err := s.db.Exec(ctx, sqlf.Sprintf(
		coordinateStartMapperQuery,
		derivativeGraphKey,
		graphKey,
		graphKey,
		graphKey,
	)); err != nil {
		return err
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(
		coordinateStartReducerQuery,
		derivativeGraphKey,
	)); err != nil {
		return err
	}

	return nil
}

const coordinateStartMapperQuery = `
INSERT INTO codeintel_ranking_progress(graph_key, max_definition_id, max_reference_id, max_path_id, mappers_started_at)
VALUES (
	%s,
	COALESCE((SELECT MAX(id) FROM codeintel_ranking_definitions WHERE graph_key = %s), 0),
	COALESCE((SELECT MAX(id) FROM codeintel_ranking_references  WHERE graph_key = %s), 0),
	COALESCE((SELECT MAX(id) FROM codeintel_initial_path_ranks  WHERE graph_key = %s), 0),
	NOW()
)
ON CONFLICT DO NOTHING
`

const coordinateStartReducerQuery = `
UPDATE codeintel_ranking_progress
SET reducer_started_at = NOW()
WHERE
	graph_key = %s AND
	mapper_completed_at IS NOT NULL AND
	seed_mapper_completed_at IS NOT NULL AND
	reducer_started_at IS NULL
`

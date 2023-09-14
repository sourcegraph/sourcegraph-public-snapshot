package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func DerivativeGraphKey(ctx context.Context, store Store) (string, time.Time, error) {
	if key, createdAt, ok, err := store.DerivativeGraphKey(ctx); err != nil {
		return "", time.Time{}, err
	} else if ok {
		return key, createdAt, nil
	}

	if err := store.BumpDerivativeGraphKey(ctx); err != nil {
		return "", time.Time{}, err
	}

	return DerivativeGraphKey(ctx, store)
}

func (s *store) DerivativeGraphKey(ctx context.Context) (graphKey string, createdAt time.Time, _ bool, err error) {
	ctx, _, endObservation := s.operations.derivativeGraphKey.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(derivativeGraphKeyQuery))
	if err != nil {
		return "", time.Time{}, false, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&graphKey, &createdAt); err != nil {
			return "", time.Time{}, false, err
		}

		return graphKey, createdAt, true, nil
	}

	return "", time.Time{}, false, nil
}

const derivativeGraphKeyQuery = `
SELECT graph_key, created_at
FROM codeintel_ranking_graph_keys
ORDER BY created_at DESC
LIMIT 1
`

// MaxGraphKeyRecords is the maximum number of graph key records we'll track before pruning older entries.
const MaxGraphKeyRecords = 10

func (s *store) BumpDerivativeGraphKey(ctx context.Context) (err error) {
	ctx, _, endObservation := s.operations.bumpDerivativeGraphKey.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(bumpDerivativeGraphKeyQuery, uuid.NewString())); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(bumpDerivativeGraphKeyPruneQuery, MaxGraphKeyRecords)); err != nil {
		return err
	}

	return nil
}

const bumpDerivativeGraphKeyQuery = `
INSERT INTO codeintel_ranking_graph_keys (graph_key) VALUES (%s)
`

const bumpDerivativeGraphKeyPruneQuery = `
DELETE FROM codeintel_ranking_graph_keys WHERE id IN (
	SELECT id
	FROM codeintel_ranking_graph_keys
	ORDER BY created_at DESC
	OFFSET %s
)
`

func (s *store) DeleteRankingProgress(ctx context.Context, graphKey string) (err error) {
	ctx, _, endObservation := s.operations.deleteRankingProgress.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(deleteRankingProgress, graphKey))
}

const deleteRankingProgress = `
DELETE FROM codeintel_ranking_progress WHERE graph_key = %s
`

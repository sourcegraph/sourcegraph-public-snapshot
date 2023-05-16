package store

import (
	"context"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) InsertReferencesForRanking(
	ctx context.Context,
	rankingGraphKey string,
	batchSize int,
	exportedUploadID int,
	references chan string,
) (err error) {
	ctx, _, endObservation := s.operations.insertReferencesForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.withTransaction(ctx, func(tx *store) error {
		inserter := func(inserter *batch.Inserter) error {
			for symbols := range batchChannel(references, batchSize) {
				if err := inserter.Insert(ctx, exportedUploadID, pq.Array(symbols), rankingGraphKey); err != nil {
					return err
				}
			}

			return nil
		}

		if err := batch.WithInserter(
			ctx,
			tx.db.Handle(),
			"codeintel_ranking_references",
			batch.MaxNumPostgresParameters,
			[]string{
				"exported_upload_id",
				"symbol_names",
				"graph_key",
			},
			inserter,
		); err != nil {
			return err
		}

		return nil
	})
}

package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) InsertInitialPathRanks(ctx context.Context, exportedUploadID int, documentPaths chan string, batchSize int, graphKey string) (err error) {
	ctx, _, endObservation := s.operations.insertInitialPathRanks.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("graphKey", graphKey),
	}})
	defer endObservation(1, observation.Args{})

	return s.withTransaction(ctx, func(tx *store) error {
		inserter := func(inserter *batch.Inserter) error {
			for paths := range batchChannel(documentPaths, batchSize) {
				if err := inserter.Insert(ctx, pq.Array(paths)); err != nil {
					return err
				}
			}

			return nil
		}

		if err := tx.db.Exec(ctx, sqlf.Sprintf(createInitialPathTemporaryTableQuery)); err != nil {
			return err
		}

		if err := batch.WithInserter(
			ctx,
			tx.db.Handle(),
			"t_codeintel_initial_path_ranks",
			batch.MaxNumPostgresParameters,
			[]string{"document_paths"},
			inserter,
		); err != nil {
			return err
		}

		if err = tx.db.Exec(ctx, sqlf.Sprintf(insertInitialPathRankCountsQuery, exportedUploadID, graphKey)); err != nil {
			return err
		}

		return nil
	})
}

const createInitialPathTemporaryTableQuery = `
CREATE TEMPORARY TABLE IF NOT EXISTS t_codeintel_initial_path_ranks (
	document_paths text[] NOT NULL
)
ON COMMIT DROP
`

const insertInitialPathRankCountsQuery = `
INSERT INTO codeintel_initial_path_ranks (exported_upload_id, document_paths, graph_key)
SELECT %s, document_paths, %s FROM t_codeintel_initial_path_ranks
`

package store

import (
	"context"
	"crypto/md5"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var sentinelPathDefinitionName = func() [16]byte {
	// This value is represented in the `insertInitialPathCountsInputsQuery`
	// Postgres query by `'\xc3e97dd6e97fb5125688c97f36720cbe'::bytea`.
	return md5.Sum([]byte("$"))
}()

func (s *store) InsertInitialPathRanks(ctx context.Context, exportedUploadID int, documentPaths []string, batchSize int, graphKey string) (err error) {
	ctx, _, endObservation := s.operations.insertInitialPathRanks.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("graphKey", graphKey),
	}})
	defer endObservation(1, observation.Args{})

	return s.withTransaction(ctx, func(tx *store) error {
		pathDefinitions, err := func() (chan shared.RankingDefinitions, error) {
			pathDefinitions := make(chan shared.RankingDefinitions, len(documentPaths))
			defer close(pathDefinitions)

			inserter := func(inserter *batch.Inserter) error {
				for _, paths := range batchSlice(documentPaths, batchSize) {
					if err := inserter.Insert(ctx, pq.Array(paths)); err != nil {
						return err
					}

					for _, path := range paths {
						pathDefinitions <- shared.RankingDefinitions{
							ExportedUploadID: exportedUploadID,
							SymbolChecksum:   sentinelPathDefinitionName,
							DocumentPath:     path,
						}
					}
				}

				return nil
			}

			if err := tx.db.Exec(ctx, sqlf.Sprintf(createInitialPathTemporaryTableQuery)); err != nil {
				return nil, err
			}

			if err := batch.WithInserter(
				ctx,
				tx.db.Handle(),
				"t_codeintel_initial_path_ranks",
				batch.MaxNumPostgresParameters,
				[]string{"document_paths"},
				inserter,
			); err != nil {
				return nil, err
			}

			if err = tx.db.Exec(ctx, sqlf.Sprintf(insertInitialPathRankCountsQuery, exportedUploadID, graphKey)); err != nil {
				return nil, err
			}

			return pathDefinitions, nil
		}()
		if err != nil {
			return err
		}

		if err := tx.InsertDefinitionsForRanking(ctx, graphKey, pathDefinitions); err != nil {
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

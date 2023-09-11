package store

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) InsertDefinitionsForRanking(
	ctx context.Context,
	rankingGraphKey string,
	definitions chan shared.RankingDefinitions,
) (err error) {
	ctx, _, endObservation := s.operations.insertDefinitionsForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.withTransaction(ctx, func(tx *store) error {
		inserter := func(inserter *batch.Inserter) error {
			for definition := range definitions {
				if err := inserter.Insert(ctx, definition.ExportedUploadID, "", derefChecksum(definition.SymbolChecksum), definition.DocumentPath, rankingGraphKey); err != nil {
					return err
				}
			}

			return nil
		}

		if err := batch.WithInserter(
			ctx,
			tx.db.Handle(),
			"codeintel_ranking_definitions",
			batch.MaxNumPostgresParameters,
			[]string{
				"exported_upload_id",
				"symbol_name",
				"symbol_checksum",
				"document_path",
				"graph_key",
			},
			inserter,
		); err != nil {
			return err
		}

		return nil
	})
}

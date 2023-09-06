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
	references chan [16]byte,
) (err error) {
	ctx, _, endObservation := s.operations.insertReferencesForRanking.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.withTransaction(ctx, func(tx *store) error {
		inserter := func(inserter *batch.Inserter) error {
			for checksums := range batchChannel(references, batchSize) {
				if err := inserter.Insert(ctx, exportedUploadID, pq.Array([]string{}), pq.Array(derefChecksums(checksums)), rankingGraphKey); err != nil {
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
				"symbol_checksums",
				"graph_key",
			},
			inserter,
		); err != nil {
			return err
		}

		return nil
	})
}

// DO NOT INLINE, the output of these functions are used by background routines
// in the batch inserter. Inlining these methods may cause a hard-to-find alias
// unless it's done very carefully.
func derefChecksums(arrs [][16]byte) [][]byte {
	cs := make([][]byte, 0, len(arrs))
	for _, arr := range arrs {
		cs = append(cs, derefChecksum(arr))
	}

	return cs
}

// DO NOT INLINE, the output of these functions are used by background routines
// in the batch inserter. Inlining these methods may cause a hard-to-find alias
// unless it's done very carefully.
func derefChecksum(arr [16]byte) []byte {
	c := make([]byte, 16)
	copy(c, arr[:])
	return c
}

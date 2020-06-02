package v5

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

func reencodeResultChunks(ctx context.Context, s *store.Store, deserializer, serializer serialization.Serializer) error {
	if err := s.Exec(ctx, sqlf.Sprintf(`CREATE TABLE "t_result_chunks" ("id" integer PRIMARY KEY NOT NULL, "data" blob NOT NULL)`)); err != nil {
		return err
	}

	rows, err := s.Query(ctx, sqlf.Sprintf("SELECT id, data FROM result_chunks"))
	if err != nil {
		return err
	}
	defer func() {
		err = store.CloseRows(rows, err)
	}()

	inserter := sqliteutil.NewBatchInserter(s, "t_result_chunks", "id", "data")

	for rows.Next() {
		var id int
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			return err
		}

		resultChunk, err := deserializer.UnmarshalResultChunkData(data)
		if err != nil {
			return err
		}

		newData, err := serializer.MarshalResultChunkData(resultChunk)
		if err != nil {
			return err
		}

		if err := inserter.Insert(ctx, id, newData); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	return swapTables(ctx, s, "result_chunks", "t_result_chunks")
}

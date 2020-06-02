package v5

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

func reencodeDocuments(ctx context.Context, s *store.Store, deserializer, serializer serialization.Serializer) error {
	if err := s.Exec(ctx, sqlf.Sprintf(`CREATE TABLE "t_documents" ("path" text PRIMARY KEY NOT NULL, "data" blob NOT NULL)`)); err != nil {
		return err
	}

	rows, err := s.Query(ctx, sqlf.Sprintf("SELECT path, data FROM documents"))
	if err != nil {
		return err
	}
	defer func() {
		err = store.CloseRows(rows, err)
	}()

	inserter := sqliteutil.NewBatchInserter(s, "t_documents", "path", "data")

	for rows.Next() {
		var path string
		var data []byte
		if err := rows.Scan(&path, &data); err != nil {
			return err
		}

		document, err := deserializer.UnmarshalDocumentData(data)
		if err != nil {
			return err
		}

		newData, err := serializer.MarshalDocumentData(document)
		if err != nil {
			return err
		}

		if err := inserter.Insert(ctx, path, newData); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	return swapTables(ctx, s, "documents", "t_documents")
}

package v5

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization/json"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

type MigrationStep func(context.Context, *store.Store, serialization.Serializer, serialization.Serializer) error

// Migrate v5: updates the serialization format of data blobs in the document, result chunks, definition,
// and reference tables. For each table, we create a new temporary table with the same columns, scan the
// rows in the original table, decode them with the old serializer, re-encode them with the new serializer,
// then write the new rows into the temporary table. The original table is dropped and the temporary table
// renamed to take its place.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	steps := []MigrationStep{
		reencodeDocuments,
		reencodeResultChunks,
		reencodeDefinitions,
		reencodeReferences,
	}

	// NOTE: We need to serialize with the JSON serializer, NOT the current serializer. This is
	// because future migrations assume that v4 was written with the most current serializer at
	// that time. Using the current serializer will cause future migrations to fail to read the
	// encoded data.
	deserializer := jsonserializer.New()

	for _, step := range steps {
		if err := step(ctx, s, deserializer, serializer); err != nil {
			return err
		}
	}

	return nil
}

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

func reencodeDefinitions(ctx context.Context, s *store.Store, deserializer, serializer serialization.Serializer) error {
	return reencodeDefinitionReferences(ctx, s, "definitions", deserializer, serializer)
}

func reencodeReferences(ctx context.Context, s *store.Store, deserializer, serializer serialization.Serializer) error {
	return reencodeDefinitionReferences(ctx, s, "references", deserializer, serializer)
}

func reencodeDefinitionReferences(ctx context.Context, s *store.Store, tableName string, deserializer, serializer serialization.Serializer) error {
	if err := s.Exec(ctx, sqlf.Sprintf(
		`CREATE TABLE "t_`+tableName+`" (
			"scheme" text NOT NULL,
			"identifier" text NOT NULL,
			"data" blob NOT NULL,
			PRIMARY KEY (scheme, identifier)
		)`,
	)); err != nil {
		return err
	}

	rows, err := s.Query(ctx, sqlf.Sprintf(`SELECT scheme, identifier, data FROM "`+tableName+`"`))
	if err != nil {
		return err
	}
	defer func() {
		err = store.CloseRows(rows, err)
	}()

	inserter := sqliteutil.NewBatchInserter(s, fmt.Sprintf("t_%s", tableName), "scheme", "identifier", "data")

	for rows.Next() {
		var scheme string
		var identifier string
		var data []byte
		if err := rows.Scan(&scheme, &identifier, &data); err != nil {
			return err
		}

		locations, err := deserializer.UnmarshalLocations(data)
		if err != nil {
			return err
		}

		newData, err := serializer.MarshalLocations(locations)
		if err != nil {
			return err
		}

		if err := inserter.Insert(ctx, scheme, identifier, newData); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	return swapTables(ctx, s, tableName, fmt.Sprintf("t_%s", tableName))
}

// swapTables deletes the original table and replaces it with the temporary table.
func swapTables(ctx context.Context, s *store.Store, tableName, tempTableName string) error {
	queries := []*sqlf.Query{
		sqlf.Sprintf(`DROP TABLE "` + tableName + `"`),
		sqlf.Sprintf(`ALTER TABLE "` + tempTableName + `" RENAME TO "` + tableName + `"`),
	}

	for _, query := range queries {
		if err := s.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

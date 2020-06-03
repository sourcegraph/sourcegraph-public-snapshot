package sqlite

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	persistence "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	gobserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization/gob"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/migrate"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

type sqliteWriter struct {
	store      *store.Store
	closer     func() error
	serializer serialization.Serializer
}

var _ persistence.Writer = &sqliteWriter{}

func NewWriter(ctx context.Context, filename string) (_ persistence.Writer, err error) {
	store, closer, err := store.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := closer(); closeErr != nil {
				err = multierror.Append(err, closeErr)
			}
		}
	}()

	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	if err := createTables(ctx, tx); err != nil {
		return nil, err
	}

	return &sqliteWriter{
		store:      tx,
		closer:     closer,
		serializer: gobserializer.New(),
	}, nil
}

func (w *sqliteWriter) WriteMeta(ctx context.Context, metaData types.MetaData) error {
	queries := []*sqlf.Query{
		sqlf.Sprintf("INSERT INTO schema_version (version) VALUES (%s)", migrate.CurrentSchemaVersion),
		sqlf.Sprintf("INSERT INTO meta (num_result_chunks) VALUES (%s)", metaData.NumResultChunks),
	}

	for _, query := range queries {
		if err := w.store.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func (w *sqliteWriter) WriteDocuments(ctx context.Context, documents map[string]types.DocumentData) error {
	inserter := sqliteutil.NewBatchInserter(w.store, "documents", "path", "data")
	for path, document := range documents {
		data, err := w.serializer.MarshalDocumentData(document)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalDocumentData")
		}

		if err := inserter.Insert(ctx, path, data); err != nil {
			return errors.Wrap(err, "documentInserter.Insert")
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return nil
}

func (w *sqliteWriter) WriteResultChunks(ctx context.Context, resultChunks map[int]types.ResultChunkData) error {
	inserter := sqliteutil.NewBatchInserter(w.store, "result_chunks", "id", "data")
	for id, resultChunk := range resultChunks {
		data, err := w.serializer.MarshalResultChunkData(resultChunk)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalResultChunkData")
		}

		if err := inserter.Insert(ctx, id, data); err != nil {
			return errors.Wrap(err, "resultChunkInserter.Insert")
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}
	return nil
}

func (w *sqliteWriter) WriteDefinitions(ctx context.Context, monikerLocations []types.MonikerLocations) error {
	return w.writeDefinitionReferences(ctx, "definitions", monikerLocations)
}

func (w *sqliteWriter) WriteReferences(ctx context.Context, monikerLocations []types.MonikerLocations) error {
	return w.writeDefinitionReferences(ctx, "references", monikerLocations)
}

func (w *sqliteWriter) writeDefinitionReferences(ctx context.Context, tableName string, monikerLocations []types.MonikerLocations) error {
	inserter := sqliteutil.NewBatchInserter(w.store, tableName, "scheme", "identifier", "data")
	for _, ml := range monikerLocations {
		data, err := w.serializer.MarshalLocations(ml.Locations)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalLocations")
		}

		if err := inserter.Insert(ctx, ml.Scheme, ml.Identifier, data); err != nil {
			return errors.Wrap(err, "inserter.Insert")
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}
	return nil
}

func (w *sqliteWriter) Close(err error) error {
	err = w.store.Done(err)

	if closeErr := w.closer(); closeErr != nil {
		err = multierror.Append(err, closeErr)
	}

	return nil
}

func createTables(ctx context.Context, store *store.Store) error {
	queries := []*sqlf.Query{
		sqlf.Sprintf(`CREATE TABLE "schema_version" ("version" text NOT NULL)`),
		sqlf.Sprintf(`CREATE TABLE "meta" ("num_result_chunks" integer NOT NULL)`),
		sqlf.Sprintf(`CREATE TABLE "documents" ("path" text PRIMARY KEY NOT NULL, "data" blob NOT NULL)`),
		sqlf.Sprintf(`CREATE TABLE "result_chunks" ("id" integer PRIMARY KEY NOT NULL, "data" blob NOT NULL)`),
		sqlf.Sprintf(`CREATE TABLE "definitions" ("scheme" text NOT NULL, "identifier" text NOT NULL, "data" blob NOT NULL, PRIMARY KEY (scheme, identifier))`),
		sqlf.Sprintf(`CREATE TABLE "references" ("scheme" text NOT NULL, "identifier" text NOT NULL, "data" blob NOT NULL, PRIMARY KEY (scheme, identifier))`),
	}

	for _, query := range queries {
		if err := store.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

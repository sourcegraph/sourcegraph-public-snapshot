package sqlite

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	persistence "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization/json"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

const InternalVersion = "0.1.0"

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

	if err := store.ExecAll(
		ctx,
		sqlf.Sprintf(`CREATE TABLE "meta" ("id" integer PRIMARY KEY NOT NULL, "lsifVersion" text NOT NULL, "sourcegraphVersion" text NOT NULL, "numResultChunks" integer NOT NULL);`),
		sqlf.Sprintf(`CREATE TABLE "documents" ("path" text PRIMARY KEY NOT NULL, "data" blob NOT NULL);`),
		sqlf.Sprintf(`CREATE TABLE "resultChunks" ("id" integer PRIMARY KEY NOT NULL, "data" blob NOT NULL);`),
		sqlf.Sprintf(`CREATE TABLE "definitions" ("id" integer PRIMARY KEY NOT NULL, "scheme" text NOT NULL, "identifier" text NOT NULL, "documentPath" text NOT NULL, "startLine" integer NOT NULL, "endLine" integer NOT NULL, "startCharacter" integer NOT NULL, "endCharacter" integer NOT NULL);`),
		sqlf.Sprintf(`CREATE TABLE "references" ("id" integer PRIMARY KEY NOT NULL, "scheme" text NOT NULL, "identifier" text NOT NULL, "documentPath" text NOT NULL, "startLine" integer NOT NULL, "endLine" integer NOT NULL, "startCharacter" integer NOT NULL, "endCharacter" integer NOT NULL);`),
	); err != nil {
		return nil, err
	}

	tx, err := store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &sqliteWriter{
		store:      tx,
		closer:     closer,
		serializer: jsonserializer.New(),
	}, nil
}

func (w *sqliteWriter) WriteMeta(ctx context.Context, meta types.MetaData) error {
	inserter := sqliteutil.NewBatchInserter(w.store, "meta", "lsifVersion", "sourcegraphVersion", "numResultChunks")
	if err := inserter.Insert(ctx, "", "", meta.NumResultChunks); err != nil {
		return errors.Wrap(err, "inserter.Insert")
	}
	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
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
	inserter := sqliteutil.NewBatchInserter(w.store, "resultChunks", "id", "data")
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
	inserter := sqliteutil.NewBatchInserter(w.store, "definitions", "scheme", "identifier", "documentPath", "startLine", "startCharacter", "endLine", "endCharacter")
	for _, ml := range monikerLocations {
		for _, l := range ml.Locations {
			if err := inserter.Insert(ctx, ml.Scheme, ml.Identifier, l.URI, l.StartLine, l.StartCharacter, l.EndLine, l.EndCharacter); err != nil {
				return errors.Wrap(err, "inserter.Insert")
			}
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return w.store.ExecAll(
		ctx,
		sqlf.Sprintf(`CREATE INDEX "idx_definitions" ON "definitions" ("scheme", "identifier")`),
	)
}

func (w *sqliteWriter) WriteReferences(ctx context.Context, monikerLocations []types.MonikerLocations) error {
	inserter := sqliteutil.NewBatchInserter(w.store, "references", "scheme", "identifier", "documentPath", "startLine", "startCharacter", "endLine", "endCharacter")
	for _, ml := range monikerLocations {
		for _, l := range ml.Locations {
			if err := inserter.Insert(ctx, ml.Scheme, ml.Identifier, l.URI, l.StartLine, l.StartCharacter, l.EndLine, l.EndCharacter); err != nil {
				return errors.Wrap(err, "inserter.Insert")
			}
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return w.store.ExecAll(
		ctx,
		sqlf.Sprintf(`CREATE INDEX "idx_references" ON "references" ("scheme", "identifier")`),
	)
}

func (w *sqliteWriter) Close() (err error) {
	err = w.store.Done(err)

	if closeErr := w.closer(); closeErr != nil {
		err = multierror.Append(err, closeErr)
	}

	return err
}

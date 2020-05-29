package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	persistence "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/schema"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

const InternalVersion = "0.1.0"

type sqliteWriter struct {
	serializer          serialization.Serializer
	db                  *sqlx.DB
	tx                  *sql.Tx
	metaInserter        *sqliteutil.BatchInserter
	documentInserter    *sqliteutil.BatchInserter
	resultChunkInserter *sqliteutil.BatchInserter
	definitionInserter  *sqliteutil.BatchInserter
	referenceInserter   *sqliteutil.BatchInserter
}

var _ persistence.Writer = &sqliteWriter{}

func NewWriter(filename string, serializer serialization.Serializer) (_ persistence.Writer, err error) {
	db, err := sqlx.Open("sqlite3_with_pcre", filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = multierror.Append(err, closeErr)
			}
		}
	}()

	if _, err := db.Exec(schema.TableDefinitions); err != nil {
		return nil, errors.Wrap(err, "creating tables")
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	metaColumns := []string{"lsifVersion", "sourcegraphVersion", "numResultChunks"}
	documentsColumns := []string{"path", "data"}
	resultChunksColumns := []string{"id", "data"}
	definitionsReferencesColumns := []string{"scheme", "identifier", "documentPath", "startLine", "startCharacter", "endLine", "endCharacter"}

	return &sqliteWriter{
		serializer:          serializer,
		db:                  db,
		tx:                  tx,
		metaInserter:        sqliteutil.NewBatchInserter(tx, "meta", metaColumns...),
		documentInserter:    sqliteutil.NewBatchInserter(tx, "documents", documentsColumns...),
		resultChunkInserter: sqliteutil.NewBatchInserter(tx, "resultChunks", resultChunksColumns...),
		definitionInserter:  sqliteutil.NewBatchInserter(tx, "definitions", definitionsReferencesColumns...),
		referenceInserter:   sqliteutil.NewBatchInserter(tx, `references`, definitionsReferencesColumns...),
	}, nil
}

func (w *sqliteWriter) WriteMeta(ctx context.Context, lsifVersion string, numResultChunks int) error {
	if err := w.metaInserter.Insert(ctx, lsifVersion, InternalVersion, numResultChunks); err != nil {
		return errors.Wrap(err, "metaInserter.Insert")
	}
	return nil
}

func (w *sqliteWriter) WriteDocuments(ctx context.Context, documents map[string]types.DocumentData) error {
	for k, v := range documents {
		ser, err := w.serializer.MarshalDocumentData(v)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalDocumentData")
		}

		if err := w.documentInserter.Insert(ctx, k, ser); err != nil {
			return errors.Wrap(err, "documentInserter.Insert")
		}
	}
	return nil
}

func (w *sqliteWriter) WriteResultChunks(ctx context.Context, resultChunks map[int]types.ResultChunkData) error {
	for k, v := range resultChunks {
		ser, err := w.serializer.MarshalResultChunkData(v)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalResultChunkData")
		}

		if err := w.resultChunkInserter.Insert(ctx, k, ser); err != nil {
			return errors.Wrap(err, "resultChunkInserter.Insert")
		}
	}
	return nil
}

func (w *sqliteWriter) WriteDefinitions(ctx context.Context, definitions []types.DefinitionReferenceRow) error {
	for _, r := range definitions {
		if err := w.definitionInserter.Insert(ctx, r.Scheme, r.Identifier, r.URI, r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter); err != nil {
			return errors.Wrap(err, "definitionInserter.Insert")
		}
	}
	return nil
}

func (w *sqliteWriter) WriteReferences(ctx context.Context, references []types.DefinitionReferenceRow) error {
	for _, r := range references {
		if err := w.referenceInserter.Insert(ctx, r.Scheme, r.Identifier, r.URI, r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter); err != nil {
			return errors.Wrap(err, "referenceInserter.Insert")
		}
	}
	return nil
}

func (w *sqliteWriter) Flush(ctx context.Context) error {
	inserters := map[string]*sqliteutil.BatchInserter{
		"metaInserter":        w.metaInserter,
		"documentInserter":    w.documentInserter,
		"resultChunkInserter": w.resultChunkInserter,
		"definitionInserter":  w.definitionInserter,
		"referenceInserter":   w.referenceInserter,
	}

	for name, inserter := range inserters {
		if err := inserter.Flush(ctx); err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s.Flush", name))
		}
	}

	if err := w.tx.Commit(); err != nil {
		return err
	}

	if _, err := w.db.ExecContext(ctx, schema.IndexDefinitions); err != nil {
		return errors.Wrap(err, "creating indexes")
	}

	return nil
}

func (w *sqliteWriter) Close() (err error) {
	if closeErr := w.db.Close(); closeErr != nil {
		err = multierror.Append(err, closeErr)
	}

	return err
}

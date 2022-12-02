package codeintel

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
)

type scipWriter struct {
	tx       *basestore.Store
	inserter *batch.Inserter
	uploadID int
}

// makeSCIPWriter creates a small wrapper over batch inserts of SCIP data. Each document
// should be written to Postgres by calling Write. The Flush method should be called after
// each document has been processed.
func makeSCIPWriter(ctx context.Context, tx *basestore.Store, uploadID int) (*scipWriter, error) {
	if err := tx.Exec(ctx, sqlf.Sprintf(makeSCIPWriterTemporaryTableQuery)); err != nil {
		return nil, err
	}

	inserter := batch.NewInserter(
		ctx,
		tx.Handle(),
		"t_codeintel_scip_symbols",
		batch.MaxNumPostgresParameters,
		"document_lookup_id",
		"symbol_name",
		"definition_ranges",
		"reference_ranges",
		"implementation_ranges",
	)

	return &scipWriter{
		tx:       tx,
		inserter: inserter,
		uploadID: uploadID,
	}, nil
}

const makeSCIPWriterTemporaryTableQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbols (
	symbol_name text NOT NULL,
	document_lookup_id integer NOT NULL,
	definition_ranges bytea,
	reference_ranges bytea,
	implementation_ranges bytea
) ON COMMIT DROP
`

// Write inserts a new document and document lookup row, and pushes all of the given
// symbols into the batch inserter.
func (s *scipWriter) Write(
	ctx context.Context,
	uploadID int,
	path string,
	payload []byte,
	payloadHash []byte,
	symbols []types.InvertedRangeIndex,
) error {
	uniquePrefix := []byte(fmt.Sprintf(
		"lsif-%d:%d:",
		uploadID,
		time.Now().UnixNano()/int64(time.Millisecond)),
	)

	documentLookupID, _, err := basestore.ScanFirstInt(s.tx.Query(ctx, sqlf.Sprintf(
		scipWriterWriteDocumentQuery,
		append(uniquePrefix, payloadHash...),
		payload,
		uploadID,
		path,
	)))
	if err != nil {
		return err
	}

	for _, symbol := range symbols {
		definitionRanges, err := types.EncodeRanges(symbol.DefinitionRanges)
		if err != nil {
			return err
		}
		referenceRanges, err := types.EncodeRanges(symbol.ReferenceRanges)
		if err != nil {
			return err
		}
		implementationRanges, err := types.EncodeRanges(symbol.ImplementationRanges)
		if err != nil {
			return err
		}

		if err := s.inserter.Insert(
			ctx,
			documentLookupID,
			symbol.SymbolName,
			definitionRanges,
			referenceRanges,
			implementationRanges,
		); err != nil {
			return err
		}
	}

	return nil
}

const scipWriterWriteDocumentQuery = `
WITH
new_document AS (
	INSERT INTO codeintel_scip_documents (schema_version, payload_hash, raw_scip_payload)
	VALUES (1, %s, %s)
	RETURNING id
)
INSERT INTO codeintel_scip_document_lookup (upload_id, document_path, document_id)
SELECT %s, %s, id FROM new_document
RETURNING id
`

// Flush ensures that all symbol writes have hit the database, and then moves all of the
// rows from the temporary table into the permanent one.
func (s *scipWriter) Flush(ctx context.Context) error {
	if err := s.inserter.Flush(ctx); err != nil {
		return err
	}

	if err := s.tx.Exec(ctx, sqlf.Sprintf(scipWriterFlushQuery, s.uploadID)); err != nil {
		return err
	}

	return nil
}

const scipWriterFlushQuery = `
INSERT INTO codeintel_scip_symbols (
	upload_id,
	symbol_name,
	document_lookup_id,
	schema_version,
	definition_ranges,
	reference_ranges,
	implementation_ranges
)
SELECT
	%s,
	source.symbol_name,
	source.document_lookup_id,
	1,
	source.definition_ranges,
	source.reference_ranges,
	source.implementation_ranges
FROM t_codeintel_scip_symbols source
ON CONFLICT DO NOTHING
`

// hashPayload returns a sha256 checksum of the given payload.
func hashPayload(payload []byte) []byte {
	hash := sha256.New()
	_, _ = hash.Write(payload)
	return hash.Sum(nil)
}

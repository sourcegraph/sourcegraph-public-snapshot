package lsifstore

import (
	"context"
	"encoding/base64"
	"sync/atomic"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type ProcessedSCIPDocument struct {
	DocumentPath   string
	Hash           []byte
	RawSCIPPayload []byte
	Symbols        []ProcessedSymbolData
	Err            error
}

type ProcessedSymbolData struct {
	SymbolName           string
	DefinitionRanges     []int32
	ReferenceRanges      []int32
	ImplementationRanges []int32
	TypeDefinitionRanges []int32
}

type ProcessedSCIPData struct {
	Documents         <-chan ProcessedSCIPDocument
	Packages          []precise.Package
	PackageReferences []precise.PackageReference
}

func (s *store) InsertSCIPDocument(ctx context.Context, uploadID int, documentPath string, hash []byte, rawSCIPPayload []byte) (_ int, err error) {
	ctx, _, endObservation := s.operations.insertSCIPDocument.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("uploadID", uploadID),
		otlog.String("documentPath", documentPath),
		otlog.String("hash", base64.StdEncoding.EncodeToString(hash)),
		otlog.Int("rawSCIPPayloadLen", len(rawSCIPPayload)),
	}})
	defer endObservation(1, observation.Args{})

	id, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		insertSCIPDocumentQuery,
		hash,
		rawSCIPPayload,
		hash,
		uploadID,
		documentPath,
	)))
	if err != nil {
		return 0, err
	}

	return id, nil
}

const insertSCIPDocumentQuery = `
WITH
new_shared_document AS (
	INSERT INTO codeintel_scip_documents (schema_version, payload_hash, raw_scip_payload)
	VALUES (1, %s, %s)
	ON CONFLICT DO NOTHING
	RETURNING id
),
shared_document AS (
	SELECT id FROM new_shared_document
	UNION ALL
	SELECT id FROM codeintel_scip_documents WHERE payload_hash = %s
)
INSERT INTO codeintel_scip_document_lookup (upload_id, document_path, document_id)
SELECT %s, %s, id FROM shared_document LIMIT 1
RETURNING id
`

func (s *store) WriteSCIPSymbols(ctx context.Context, uploadID, documentLookupID int, symbols []ProcessedSymbolData) (count uint32, err error) {
	ctx, trace, endObservation := s.operations.writeSCIPSymbols.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("uploadID", uploadID),
		otlog.Int("documentLookupID", documentLookupID),
		otlog.Int("numSymbols", len(symbols)),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(writeSCIPSymbolsTemporaryTableQuery)); err != nil {
		return 0, err
	}

	inserter := func(inserter *batch.Inserter) error {
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
			typeDefinitionRanges, err := types.EncodeRanges(symbol.TypeDefinitionRanges)
			if err != nil {
				return err
			}

			if err := inserter.Insert(
				ctx,
				uploadID,
				symbol.SymbolName,
				definitionRanges,
				referenceRanges,
				implementationRanges,
				typeDefinitionRanges,
			); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}

		return nil
	}

	if err := withSingleThreadedBatchInserter(
		ctx,
		tx.Handle(),
		"t_codeintel_scip_symbols",
		[]string{
			"upload_id",
			"symbol_name",
			"definition_ranges",
			"reference_ranges",
			"implementation_ranges",
			"type_definition_ranges",
		},
		inserter,
	); err != nil {
		return 0, err
	}
	trace.Log(log.Int("numRecords", int(count)))

	err = tx.Exec(ctx, sqlf.Sprintf(writeSCIPSymbolsInsertQuery, documentLookupID, 1))
	if err != nil {
		return 0, err
	}

	return count, nil
}

const writeSCIPSymbolsTemporaryTableQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbols (
	upload_id integer NOT NULL,
	symbol_name text NOT NULL,
	definition_ranges bytea,
	reference_ranges bytea,
	implementation_ranges bytea,
	type_definition_ranges bytea
) ON COMMIT DROP
`

const writeSCIPSymbolsInsertQuery = `
INSERT INTO codeintel_scip_symbols (
	upload_id,
	symbol_name,
	document_lookup_id,
	schema_version,
	definition_ranges,
	reference_ranges,
	implementation_ranges,
	type_definition_ranges
)
SELECT
	source.upload_id,
	source.symbol_name,
	%s,
	%s,
	source.definition_ranges,
	source.reference_ranges,
	source.implementation_ranges,
	source.type_definition_ranges
FROM t_codeintel_scip_symbols source
ON CONFLICT DO NOTHING
`

func withSingleThreadedBatchInserter(ctx context.Context, db dbutil.DB, tableName string, columns []string, f func(inserter *batch.Inserter) error) (err error) {
	return batch.WithInserter(ctx, db, tableName, batch.MaxNumPostgresParameters, columns, f)
}

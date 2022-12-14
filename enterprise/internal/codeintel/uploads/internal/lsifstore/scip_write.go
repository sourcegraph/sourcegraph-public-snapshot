package lsifstore

import (
	"context"
	"sort"
	"sync/atomic"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/trie"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ProcessedSCIPData struct {
	Metadata          ProcessedMetadata
	Documents         <-chan ProcessedSCIPDocument
	Packages          <-chan precise.Package
	PackageReferences <-chan precise.PackageReference
}

type ProcessedMetadata struct {
	TextDocumentEncoding string
	ToolName             string
	ToolVersion          string
	ToolArguments        []string
	ProtocolVersion      int
}

type ProcessedSCIPDocument struct {
	DocumentPath   string
	Hash           []byte
	RawSCIPPayload []byte
	Symbols        []types.InvertedRangeIndex
	Err            error
}

func (s *store) InsertMetadata(ctx context.Context, uploadID int, meta ProcessedMetadata) (err error) {
	ctx, _, endObservation := s.operations.insertMetadata.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	if meta.ToolArguments == nil {
		meta.ToolArguments = []string{}
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(
		insertMetadataQuery,
		uploadID,
		meta.TextDocumentEncoding,
		meta.ToolName,
		meta.ToolVersion,
		pq.Array(meta.ToolArguments),
		meta.ProtocolVersion,
	)); err != nil {
		return err
	}

	return nil
}

const insertMetadataQuery = `
INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version)
VALUES (%s, %s, %s, %s, %s, %s)
`

func (s *store) NewSCIPWriter(ctx context.Context, uploadID int) (SCIPWriter, error) {
	if !s.db.InTransaction() {
		return nil, errors.New("WriteSCIPSymbols must be called in a transaction")
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(writeSCIPSymbolsTemporarySymbolNamesTableQuery)); err != nil {
		return nil, err
	}
	if err := s.db.Exec(ctx, sqlf.Sprintf(writeSCIPSymbolsTemporarySymbolsTableQuery)); err != nil {
		return nil, err
	}

	symbolNameInserter := batch.NewInserter(
		ctx,
		s.db.Handle(),
		"t_codeintel_scip_symbol_names",
		batch.MaxNumPostgresParameters,
		"id",
		"name_segment",
		"prefix_id",
	)

	symbolInserter := batch.NewInserter(
		ctx,
		s.db.Handle(),
		"t_codeintel_scip_symbols",
		batch.MaxNumPostgresParameters,
		"document_lookup_id",
		"symbol_id",
		"definition_ranges",
		"reference_ranges",
		"implementation_ranges",
		"type_definition_ranges",
	)

	scipWriter := &scipWriter{
		uploadID:           uploadID,
		db:                 s.db,
		symbolNameInserter: symbolNameInserter,
		symbolInserter:     symbolInserter,
		count:              0,
	}

	return scipWriter, nil
}

const writeSCIPSymbolsTemporarySymbolNamesTableQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbol_names (
	id integer NOT NULL,
	name_segment text NOT NULL,
	prefix_id integer
) ON COMMIT DROP
`

const writeSCIPSymbolsTemporarySymbolsTableQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbols (
	symbol_id integer NOT NULL,
	document_lookup_id integer NOT NULL,
	definition_ranges bytea,
	reference_ranges bytea,
	implementation_ranges bytea,
	type_definition_ranges bytea
) ON COMMIT DROP
`

type scipWriter struct {
	uploadID           int
	nextID             int
	db                 *basestore.Store
	symbolNameInserter *batch.Inserter
	symbolInserter     *batch.Inserter
	count              uint32
}

func (s *scipWriter) InsertDocument(ctx context.Context, documentPath string, hash []byte, rawSCIPPayload []byte, symbols []types.InvertedRangeIndex) error {
	documentLookupID, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		insertSCIPDocumentQuery,
		1,
		hash,
		rawSCIPPayload,
		hash,
		s.uploadID,
		documentPath,
	)))
	if err != nil {
		return err
	}

	return s.WriteSCIPSymbols(ctx, documentLookupID, symbols)
}

const insertSCIPDocumentQuery = `
WITH
new_shared_document AS (
	INSERT INTO codeintel_scip_documents (schema_version, payload_hash, raw_scip_payload)
	VALUES (%s, %s, %s)
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

func (s *scipWriter) WriteSCIPSymbols(ctx context.Context, documentLookupID int, symbols []types.InvertedRangeIndex) error {
	symbolNameMap := make(map[string]struct{}, len(symbols))
	for _, invertedRange := range symbols {
		symbolNameMap[invertedRange.SymbolName] = struct{}{}
	}
	symbolNames := make([]string, 0, len(symbolNameMap))
	for symbolName := range symbolNameMap {
		symbolNames = append(symbolNames, symbolName)
	}
	sort.Strings(symbolNames)

	var symbolNameTrie trie.Trie
	symbolNameTrie, s.nextID = trie.NewTrie(symbolNames, s.nextID)

	symbolNameByIDs := map[int]string{}
	idsBySymbolName := map[string]int{}

	if err := symbolNameTrie.Traverse(func(id int, parentID *int, prefix string) error {
		name := prefix
		if parentID != nil {
			parentPrefix, ok := symbolNameByIDs[*parentID]
			if !ok {
				return errors.Newf("malformed trie - expected prefix with id=%d to exist", *parentID)
			}

			name = parentPrefix + prefix
		}
		symbolNameByIDs[id] = name
		idsBySymbolName[name] = id

		if err := s.symbolNameInserter.Insert(ctx, id, prefix, parentID); err != nil {
			return err
		}

		return nil
	}); err != nil {
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
		typeDefinitionRanges, err := types.EncodeRanges(symbol.TypeDefinitionRanges)
		if err != nil {
			return err
		}

		symbolID, ok := idsBySymbolName[symbol.SymbolName]
		if !ok {
			return errors.Newf("malformed trie - expected %q to be a member", symbol.SymbolName)
		}

		if err := s.symbolInserter.Insert(
			ctx,
			documentLookupID,
			symbolID,
			definitionRanges,
			referenceRanges,
			implementationRanges,
			typeDefinitionRanges,
		); err != nil {
			return err
		}

		atomic.AddUint32(&s.count, 1)
	}

	return nil
}

func (s *scipWriter) flush(ctx context.Context) error {
	// TODO
	return nil
}

func (s *scipWriter) Flush(ctx context.Context) (uint32, error) {
	// Flush all buffered documents
	if err := s.flush(ctx); err != nil {
		return 0, err
	}

	// Flush all data into temp tables
	if err := s.symbolNameInserter.Flush(ctx); err != nil {
		return 0, err
	}
	if err := s.symbolInserter.Flush(ctx); err != nil {
		return 0, err
	}

	// Move all data from temp tables into target tables
	if err := s.db.Exec(ctx, sqlf.Sprintf(writeSCIPSymbolsFlushSymbolNamesQuery, s.uploadID)); err != nil {
		return 0, err
	}
	if err := s.db.Exec(ctx, sqlf.Sprintf(writeSCIPSymbolsFlushSymbolsQuery, s.uploadID, 1)); err != nil {
		return 0, err
	}

	return s.count, nil
}

const writeSCIPSymbolsFlushSymbolNamesQuery = `
INSERT INTO codeintel_scip_symbol_names (
	upload_id,
	id,
	name_segment,
	prefix_id
)
SELECT
	%s,
	source.id,
	source.name_segment,
	source.prefix_id
FROM t_codeintel_scip_symbol_names source
`

const writeSCIPSymbolsFlushSymbolsQuery = `
INSERT INTO codeintel_scip_symbols (
	upload_id,
	symbol_id,
	document_lookup_id,
	schema_version,
	definition_ranges,
	reference_ranges,
	implementation_ranges,
	type_definition_ranges
)
SELECT
	%s,
	source.symbol_id,
	source.document_lookup_id,
	%s,
	source.definition_ranges,
	source.reference_ranges,
	source.implementation_ranges,
	source.type_definition_ranges
FROM t_codeintel_scip_symbols source
`

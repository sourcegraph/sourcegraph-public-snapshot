package lsifstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync/atomic"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/ranges"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/symbols"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO - move
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
	Path     string
	Document *scip.Document
	Err      error
}

func (s *store) InsertMetadata(ctx context.Context, uploadID int, meta ProcessedMetadata) (err error) {
	ctx, _, endObservation := s.operations.insertMetadata.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("uploadID", uploadID),
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

	if err := s.db.Exec(ctx, sqlf.Sprintf(newSCIPWriterTemporarySymbolsTableQuery)); err != nil {
		return nil, err
	}
	if err := s.db.Exec(ctx, sqlf.Sprintf(newSCIPWriterTemporarySymbolLookupTableQuery)); err != nil {
		return nil, err
	}

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

		"scheme_id",
		"package_manager_id",
		"package_name_id",
		"package_version_id",
		"descriptor_id",
		"descriptor_no_suffix_id",
	)

	symbolLookupInserter := batch.NewInserter(
		ctx,
		s.db.Handle(),
		"t_codeintel_scip_symbols_lookup",
		batch.MaxNumPostgresParameters,
		"id",
		"upload_id",
		"name",
		"scip_name_type",
	)

	scipWriter := &scipWriter{
		uploadID:             uploadID,
		db:                   s.db,
		symbolInserter:       symbolInserter,
		symbolLookupInserter: symbolLookupInserter,
		count:                0,
	}

	return scipWriter, nil
}

const newSCIPWriterTemporarySymbolsTableQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbols (
	symbol_id integer NOT NULL,
	document_lookup_id integer NOT NULL,
	definition_ranges bytea,
	reference_ranges bytea,
	implementation_ranges bytea,
	type_definition_ranges bytea,

	scheme_id integer,
	package_manager_id integer,
	package_name_id integer,
	package_version_id integer,
	descriptor_id integer,
	descriptor_no_suffix_id integer
) ON COMMIT DROP
`

const newSCIPWriterTemporarySymbolLookupTableQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbols_lookup(
	id integer NOT NULL,
	upload_id integer NOT NULL,
	name text NOT NULL,
	scip_name_type text NOT NULL
) ON COMMIT DROP
`

type scipWriter struct {
	uploadID             int
	nextSymbolLookupID   int
	nextSymbolID         int
	db                   *basestore.Store
	symbolInserter       *batch.Inserter
	symbolLookupInserter *batch.Inserter
	count                uint32
	batchPayloadSum      int
	batch                []bufferedDocument
}

type bufferedDocument struct {
	path         string
	scipDocument *scip.Document
	payload      []byte
	payloadHash  []byte
}

const (
	DocumentsBatchSize = 256
	MaxBatchPayloadSum = 1024 * 1024 * 32
)

func (s *scipWriter) InsertDocument(ctx context.Context, path string, scipDocument *scip.Document) error {
	if s.batchPayloadSum >= MaxBatchPayloadSum {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	payload, err := proto.Marshal(scipDocument)
	if err != nil {
		return err
	}

	compressedPayload, err := shared.Compressor.Compress(bytes.NewReader(payload))
	if err != nil {
		return err
	}

	s.batch = append(s.batch, bufferedDocument{
		path:         path,
		scipDocument: scipDocument,
		payload:      compressedPayload,
		payloadHash:  hashPayload(payload),
	})
	s.batchPayloadSum += len(payload)

	if len(s.batch) >= DocumentsBatchSize {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *scipWriter) flush(ctx context.Context) error {
	documents := s.batch
	s.batch = nil
	s.batchPayloadSum = 0

	documentIDs, err := batch.WithInserterForIdentifiers(
		ctx,
		s.db.Handle(),
		"codeintel_scip_documents",
		batch.MaxNumPostgresParameters,
		[]string{
			"schema_version",
			"payload_hash",
			"raw_scip_payload",
		},
		"ON CONFLICT DO NOTHING",
		"id",
		func(inserter *batch.Inserter) error {
			for _, document := range documents {
				if err := inserter.Insert(ctx, 1, document.payloadHash, document.payload); err != nil {
					return err
				}
			}

			return nil
		},
	)
	if err != nil {
		return err
	}
	if len(documentIDs) != len(documents) {
		hashes := make([][]byte, 0, len(documents))
		hashSet := make(map[string]struct{}, len(documents))
		for _, document := range documents {
			key := hex.EncodeToString(document.payloadHash)
			if _, ok := hashSet[key]; !ok {
				hashSet[key] = struct{}{}
				hashes = append(hashes, document.payloadHash)
			}
		}
		idsByHash, err := scanIDsByHash(s.db.Query(ctx, sqlf.Sprintf(scipWriterWriteFetchDocumentsQuery, pq.Array(hashes))))
		if err != nil {
			return err
		}
		documentIDs = documentIDs[:0]
		for _, document := range documents {
			documentIDs = append(documentIDs, idsByHash[hex.EncodeToString(document.payloadHash)])
		}
		if len(idsByHash) != len(hashes) {
			return errors.New("unexpected number of document records inserted/retrieved")
		}
	}

	documentLookupIDs, err := batch.WithInserterForIdentifiers(
		ctx,
		s.db.Handle(),
		"codeintel_scip_document_lookup",
		batch.MaxNumPostgresParameters,
		[]string{
			"upload_id",
			"document_path",
			"document_id",
		},
		"",
		"id",
		func(inserter *batch.Inserter) error {
			for i, document := range documents {
				if err := inserter.Insert(ctx, s.uploadID, document.path, documentIDs[i]); err != nil {
					return err
				}
			}

			return nil
		},
	)
	if err != nil {
		return err
	}
	if len(documentLookupIDs) != len(documents) {
		return errors.New("unexpected number of document lookup records inserted")
	}

	symbolNameMap := map[string]struct{}{}
	invertedRangeIndexes := make([][]shared.InvertedRangeIndex, 0, len(documents))
	for _, document := range documents {
		index := shared.ExtractSymbolIndexes(document.scipDocument)
		invertedRangeIndexes = append(invertedRangeIndexes, index)

		for _, invertedRange := range index {
			symbolNameMap[invertedRange.SymbolName] = struct{}{}
		}
	}
	symbolNames := make([]string, 0, len(symbolNameMap))
	for symbolName := range symbolNameMap {
		symbolNames = append(symbolNames, symbolName)
	}
	sort.Strings(symbolNames)

	schemes := make(map[string]int)
	managers := make(map[string]int)
	packageNames := make(map[string]int)
	packageVersions := make(map[string]int)
	descriptors := make(map[string]int)
	descriptorsNoSuffix := make(map[string]int)

	type explodedIDs struct {
		schemeID             int
		packageManagerID     int
		packageNameID        int
		packageVersionID     int
		descriptorID         int
		descriptorNoSuffixID int
	}
	cache := map[string]explodedIDs{}

	getOrSetID := func(m map[string]int, key string) int {
		if v, ok := m[key]; ok {
			return v
		}

		id := s.nextSymbolLookupID
		s.nextSymbolLookupID++
		m[key] = id
		return id
	}

	for _, symbolName := range symbolNames {
		symbol, err := symbols.NewExplodedSymbol(symbolName)
		if err != nil {
			return err
		}

		schemeID := getOrSetID(schemes, symbol.Scheme)
		packageManagerID := getOrSetID(managers, symbol.PackageManager)
		packageNameID := getOrSetID(packageNames, symbol.PackageName)
		packageVersionID := getOrSetID(packageVersions, symbol.PackageVersion)
		descriptorID := getOrSetID(descriptors, symbol.Descriptor)
		descriptorNoSuffixID := getOrSetID(descriptorsNoSuffix, symbol.DescriptorNoSuffix)
		cache[symbolName] = explodedIDs{
			schemeID:             schemeID,
			packageManagerID:     packageManagerID,
			packageNameID:        packageNameID,
			packageVersionID:     packageVersionID,
			descriptorID:         descriptorID,
			descriptorNoSuffixID: descriptorNoSuffixID,
		}
	}

	maps := map[string]map[string]int{
		"SCHEME":               schemes,
		"PACKAGE_MANAGER":      managers,
		"PACKAGE_NAME":         packageNames,
		"PACKAGE_VERSION":      packageVersions,
		"DESCRIPTOR":           descriptors,
		"DESCRIPTOR_NO_SUFFIX": descriptorsNoSuffix,
	}

	for nameType, m := range maps {
		for symbolName, symbolID := range m {
			if err := s.symbolLookupInserter.Insert(ctx, symbolID, s.uploadID, symbolName, nameType); err != nil {
				return err
			}
		}
	}

	for i, invertedRangeIndexes := range invertedRangeIndexes {
		for _, index := range invertedRangeIndexes {
			definitionRanges, err := ranges.EncodeRanges(index.DefinitionRanges)
			if err != nil {
				return err
			}
			referenceRanges, err := ranges.EncodeRanges(index.ReferenceRanges)
			if err != nil {
				return err
			}
			implementationRanges, err := ranges.EncodeRanges(index.ImplementationRanges)
			if err != nil {
				return err
			}
			typeDefinitionRanges, err := ranges.EncodeRanges(index.TypeDefinitionRanges)
			if err != nil {
				return err
			}

			s.nextSymbolID++
			ids := cache[index.SymbolName]

			if err := s.symbolInserter.Insert(
				ctx,
				documentLookupIDs[i],
				s.nextSymbolID,
				definitionRanges,
				referenceRanges,
				implementationRanges,
				typeDefinitionRanges,

				ids.schemeID,
				ids.packageManagerID,
				ids.packageNameID,
				ids.packageVersionID,
				ids.descriptorID,
				ids.descriptorNoSuffixID,
			); err != nil {
				return err
			}

			atomic.AddUint32(&s.count, 1)
		}
	}

	return nil
}

const scipWriterWriteFetchDocumentsQuery = `
SELECT
	encode(payload_hash, 'hex'),
	id
FROM codeintel_scip_documents
WHERE payload_hash = ANY(%s)
`

func (s *scipWriter) Flush(ctx context.Context) (uint32, error) {
	// Flush all buffered documents
	if err := s.flush(ctx); err != nil {
		return 0, err
	}

	// Flush all data into temp tables
	if err := s.symbolLookupInserter.Flush(ctx); err != nil {
		return 0, err
	}
	if err := s.symbolInserter.Flush(ctx); err != nil {
		return 0, err
	}

	// Move all data from temp tables into target tables
	if err := s.db.Exec(ctx, sqlf.Sprintf(scipWriterFlushSymbolLookupQuery, s.uploadID)); err != nil {
		return 0, err
	}
	if err := s.db.Exec(ctx, sqlf.Sprintf(scipWriterFlushSymbolsQuery, s.uploadID, 1)); err != nil {
		return 0, err
	}

	return s.count, nil
}

const scipWriterFlushSymbolLookupQuery = `
INSERT INTO codeintel_scip_symbols_lookup (
	upload_id,
	id,
	name,
	scip_name_type
)
SELECT
	%s,
	source.id,
	source.name,
	source.scip_name_type
FROM t_codeintel_scip_symbols_lookup source
`

const scipWriterFlushSymbolsQuery = `
INSERT INTO codeintel_scip_symbols (
	upload_id,
	symbol_id,
	document_lookup_id,
	schema_version,
	definition_ranges,
	reference_ranges,
	implementation_ranges,
	type_definition_ranges,

	scheme_id,
	package_manager_id,
	package_name_id,
	package_version_id,
	descriptor_id,
	descriptor_no_suffix_id
)
SELECT
	%s,
	source.symbol_id,
	source.document_lookup_id,
	%s,
	source.definition_ranges,
	source.reference_ranges,
	source.implementation_ranges,
	source.type_definition_ranges,

	source.scheme_id,
	source.package_manager_id,
	source.package_name_id,
	source.package_version_id,
	source.descriptor_id,
	source.descriptor_no_suffix_id
FROM t_codeintel_scip_symbols source
`

// hashPayload returns a sha256 checksum of the given payload.
func hashPayload(payload []byte) []byte {
	hash := sha256.New()
	_, _ = hash.Write(payload)
	return hash.Sum(nil)
}

var scanIDsByHash = basestore.NewMapScanner(func(s dbutil.Scanner) (hash string, id int, _ error) {
	err := s.Scan(&hash, &id)
	return hash, id, err
})

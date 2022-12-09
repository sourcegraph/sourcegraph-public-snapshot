package codeintel

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"google.golang.org/protobuf/proto"
	"k8s.io/utils/lru"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/scip"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type scipMigrator struct {
	store          *basestore.Store
	codeintelStore *basestore.Store
	serializer     *serializer
}

func NewSCIPMigrator(store, codeintelStore *basestore.Store) *scipMigrator {
	return &scipMigrator{
		store:          store,
		codeintelStore: codeintelStore,
		serializer:     newSerializer(),
	}
}

func (m *scipMigrator) ID() int                 { return 18 }
func (m *scipMigrator) Interval() time.Duration { return time.Second }

// Progress returns the ratio between the number of SCIP upload records to SCIP+LSIF upload.
func (m *scipMigrator) Progress(ctx context.Context, applyReverse bool) (float64, error) {
	if applyReverse {
		// If we're applying this in reverse, just report 0% immediately. If we have any SCIP
		// records, we will lose access to them on a downgrade, but will leave them on-disk in
		// the event of a successful re-upgrade.
		return 0, nil
	}

	progress, _, err := basestore.ScanFirstFloat(m.codeintelStore.Query(ctx, sqlf.Sprintf(
		scipMigratorProgressQuery,
	)))
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const scipMigratorProgressQuery = `
SELECT CASE c1.count + c2.count WHEN 0 THEN 1 ELSE cast(c1.count as float) / cast((c1.count + c2.count) as float) END FROM
	(SELECT COUNT(*) as count FROM codeintel_scip_metadata) c1,
	(SELECT COUNT(*) as count FROM lsif_data_metadata) c2
`

var (
	// NOTE: modified in tests
	scipMigratorUploadBatchSize             = 64
	scipMigratorDocumentBatchSize           = 128
	scipMigratorResultChunkDefaultCacheSize = 1024
)

func (m *scipMigrator) Up(ctx context.Context) error {
	for i := 0; i < scipMigratorUploadBatchSize; i++ {
		if err := m.upSingle(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (m *scipMigrator) upSingle(ctx context.Context) (err error) {
	tx, err := m.codeintelStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Select an upload record to process and lock it in this transaction so that we don't
	// compete with other migrator routines that may be running.
	uploadID, ok, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(scipMigratorSelectForMigrationQuery)))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	scipWriter, err := makeSCIPWriter(ctx, tx, uploadID)
	if err != nil {
		return err
	}
	if err := migrateUpload(ctx, m.store, tx, m.serializer, scipWriter, uploadID); err != nil {
		return err
	}
	if err := scipWriter.Flush(ctx); err != nil {
		return err
	}
	if err := deleteLSIFData(ctx, tx, uploadID); err != nil {
		return err
	}

	return nil
}

const scipMigratorSelectForMigrationQuery = `
SELECT dump_id
FROM lsif_data_metadata
ORDER BY dump_id
FOR UPDATE SKIP LOCKED
LIMIT 1
`

func (m *scipMigrator) Down(ctx context.Context) error {
	// We shouldn't return > 0% on apply reverse, should not be called.
	return nil
}

// migrateUpload converts each LSIF document belonging to the given upload into a SCIP document
// and persists them to the codeintel-db in the given transaction.
func migrateUpload(
	ctx context.Context,
	store *basestore.Store,
	codeintelTx *basestore.Store,
	serializer *serializer,
	scipWriter *scipWriter,
	uploadID int,
) error {
	indexerName, _, err := basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(
		scipMigratorIndexerQuery,
		uploadID,
	)))
	if err != nil {
		return err
	}

	numResultChunks, ok, err := basestore.ScanFirstInt(codeintelTx.Query(ctx, sqlf.Sprintf(
		scipMigratorReadMetadataQuery,
		uploadID,
	)))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	cacheSize := scipMigratorResultChunkDefaultCacheSize
	if numResultChunks < cacheSize {
		cacheSize = numResultChunks
	}
	resultChunkCache := lru.New(cacheSize)

	for page := 0; ; page++ {
		documentsByPath, err := makeDocumentScanner(serializer)(codeintelTx.Query(ctx, sqlf.Sprintf(
			scipMigratorScanDocumentsQuery,
			uploadID,
			scipMigratorDocumentBatchSize,
			page*scipMigratorDocumentBatchSize,
		)))
		if err != nil {
			return err
		}
		if len(documentsByPath) == 0 {
			break
		}

		for path, document := range documentsByPath {
			if err := processDocument(
				ctx,
				codeintelTx,
				serializer,
				scipWriter,
				resultChunkCache,
				uploadID,
				numResultChunks,
				indexerName,
				path,
				document,
			); err != nil {
				return err
			}
		}
	}

	if err := codeintelTx.Exec(ctx, sqlf.Sprintf(
		scipMigratorWriteMetadataQuery,
		uploadID,
	)); err != nil {
		return err
	}

	return nil
}

const scipMigratorIndexerQuery = `
SELECT indexer
FROM lsif_uploads
WHERE id = %s
`

const scipMigratorReadMetadataQuery = `
SELECT num_result_chunks
FROM lsif_data_metadata
WHERE dump_id = %s
`

const scipMigratorScanDocumentsQuery = `
SELECT
	path,
	ranges,
	hovers,
	monikers,
	packages,
	diagnostics
FROM lsif_data_documents
WHERE dump_id = %s
ORDER BY path
LIMIT %s
OFFSET %d
`

const scipMigratorWriteMetadataQuery = `
INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version)
VALUES (%s, '', '', '', '{}', 1)
`

// processDocument converts the given LSIF document into a SCIP document and persists it to the
// codeintel-db in the given transaction.
func processDocument(
	ctx context.Context,
	tx *basestore.Store,
	serializer *serializer,
	scipWriter *scipWriter,
	resultChunkCache *lru.Cache,
	uploadID int,
	numResultChunks int,
	indexerName,
	path string,
	document DocumentData,
) error {

	// We first read the relevant result chunks for this document into memory, writing them through to the
	// shared result chunk cache to avoid re-fetching result chunks that are used to processed to documents
	// in a row.

	resultChunks, err := fetchResultChunks(
		ctx,
		tx,
		serializer,
		resultChunkCache,
		uploadID,
		numResultChunks,
		extractDefinitionResultIDs(document.Ranges),
	)
	if err != nil {
		return err
	}

	definitionMatcher := func(
		targetPath string,
		targetRangeID precise.ID,
		definitionResultID precise.ID,
	) bool {
		definitionResultChunk, ok := resultChunks[precise.HashKey(definitionResultID, numResultChunks)]
		if !ok {
			return false
		}

		for _, pair := range definitionResultChunk.DocumentIDRangeIDs[ID(definitionResultID)] {
			if targetPath == definitionResultChunk.DocumentPaths[pair.DocumentID] && pair.RangeID == ID(targetRangeID) {
				return true
			}
		}

		return false
	}

	scipDocument := types.CanonicalizeDocument(scip.ConvertLSIFDocument(
		uploadID,
		definitionMatcher,
		indexerName,
		path,
		toPreciseTypes(document),
	))

	payload, err := proto.Marshal(scipDocument)
	if err != nil {
		return err
	}

	if err := scipWriter.Write(
		ctx,
		uploadID,
		path,
		payload,
		hashPayload(payload),
		types.ExtractSymbolIndexes(scipDocument),
	); err != nil {
		return err
	}

	return nil
}

// fetchResultChunks queries for the set of result chunks containing one of the given result set
// identifiers. The output of this function is a map from result chunk index to unmarshalled data.
func fetchResultChunks(
	ctx context.Context,
	tx *basestore.Store,
	serializer *serializer,
	resultChunkCache *lru.Cache,
	uploadID int,
	numResultChunks int,
	ids []ID,
) (map[int]ResultChunkData, error) {
	resultChunks := map[int]ResultChunkData{}
	indexMap := map[int]struct{}{}

	for _, id := range ids {
		// Calculate result chunk index that this identifier belongs to
		idx := precise.HashKey(precise.ID(id), numResultChunks)

		// Skip if we already loaded this result chunk from the cache
		if _, ok := resultChunks[idx]; ok {
			continue
		}

		// Attempt to load result chunk data from the cache. If it's present then we can add it to
		// the output map immediately. If it's not present in the cache, then we'll need to fetch it
		// from the database. Collect each such result chunk index so we can do a batch load.

		if rawResultChunk, ok := resultChunkCache.Get(idx); ok {
			resultChunks[idx] = rawResultChunk.(ResultChunkData)
		} else {
			indexMap[idx] = struct{}{}
		}
	}

	if len(indexMap) > 0 {
		indexes := make([]int, len(indexMap))
		for index := range indexMap {
			indexes = append(indexes, index)
		}

		// Fetch missing result chunks from the database. Add each of the loaded result chunks into
		// the cache shared while processing this particular upload.

		scanResultChunks := scanResultChunksIntoMap(serializer, func(idx int, resultChunk ResultChunkData) error {
			resultChunks[idx] = resultChunk
			resultChunkCache.Add(idx, resultChunk)
			return nil
		})
		if err := scanResultChunks(tx.Query(ctx, sqlf.Sprintf(
			scipMigratorScanResultChunksQuery,
			uploadID,
			pq.Array(indexes),
		))); err != nil {
			return nil, err
		}
	}

	return resultChunks, nil
}

const scipMigratorScanResultChunksQuery = `
SELECT
	idx,
	data
FROM lsif_data_result_chunks
WHERE
	dump_id = %s AND
	idx = ANY(%s)
`

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

var lsifTableNames = []string{
	"lsif_data_metadata",
	"lsif_data_documents",
	"lsif_data_result_chunks",
	"lsif_data_definitions",
	"lsif_data_references",
	"lsif_data_implementations",
}

func deleteLSIFData(ctx context.Context, tx *basestore.Store, uploadID int) error {
	for _, tableName := range lsifTableNames {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			deleteLSIFDataQuery,
			sqlf.Sprintf(tableName),
			uploadID,
		)); err != nil {
			return err
		}
	}

	return nil
}

const deleteLSIFDataQuery = `
DELETE FROM %s WHERE dump_id = %s
`

func makeDocumentScanner(serializer *serializer) func(rows *sql.Rows, queryErr error) (map[string]DocumentData, error) {
	return basestore.NewMapScanner(func(s dbutil.Scanner) (string, DocumentData, error) {
		var path string
		var data MarshalledDocumentData
		if err := s.Scan(&path, &data.Ranges, &data.HoverResults, &data.Monikers, &data.PackageInformation, &data.Diagnostics); err != nil {
			return "", DocumentData{}, err
		}

		document, err := serializer.UnmarshalDocumentData(data)
		if err != nil {
			return "", DocumentData{}, err
		}

		return path, document, nil
	})
}

func scanResultChunksIntoMap(serializer *serializer, f func(idx int, resultChunk ResultChunkData) error) func(rows *sql.Rows, queryErr error) error {
	return basestore.NewCallbackScanner(func(s dbutil.Scanner) error {
		var idx int
		var rawData []byte
		if err := s.Scan(&idx, &rawData); err != nil {
			return err
		}

		data, err := serializer.UnmarshalResultChunkData(rawData)
		if err != nil {
			return err
		}

		return f(idx, data)
	})
}

// extractDefinitionResultIDs extracts the non-empty identifiers of the LSIF definition results attached to
// any of the given ranges. The returned identifiers are unique and ordered.
func extractDefinitionResultIDs(ranges map[ID]RangeData) []ID {
	resultIDMap := map[ID]struct{}{}
	for _, r := range ranges {
		if r.DefinitionResultID != "" {
			resultIDMap[r.DefinitionResultID] = struct{}{}
		}
	}

	ids := make([]ID, 0, len(resultIDMap))
	for id := range resultIDMap {
		ids = append(ids, id)
	}
	return ids
}

// hashPayload returns a sha256 checksum of the given payload.
func hashPayload(payload []byte) []byte {
	hash := sha256.New()
	_, _ = hash.Write(payload)
	return hash.Sum(nil)
}

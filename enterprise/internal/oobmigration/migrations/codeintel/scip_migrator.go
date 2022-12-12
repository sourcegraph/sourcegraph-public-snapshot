package codeintel

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	ogscip "github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"
	"k8s.io/utils/lru"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/scip"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (m *scipMigrator) ID() int                 { return 19 }
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
	// scipMigratorUploadConcurrency           = 4
	scipMigratorUploadBatchSize             = 1
	scipMigratorDocumentBatchSize           = 32
	scipMigratorResultChunkDefaultCacheSize = 8192
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
) (err error) {
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

	// Warm result chunk cache if it will all fit in the cache
	if numResultChunks <= cacheSize {
		var ids []ID
		for i := 0; i < numResultChunks; i++ {
			ids = append(ids, ID(fmt.Sprintf("%d", i)))
		}

		scanResultChunks := scanResultChunksIntoMap(serializer, func(idx int, resultChunk ResultChunkData) error {
			resultChunkCache.Add(idx, resultChunk)
			return nil
		})
		if err := scanResultChunks(codeintelTx.Query(ctx, sqlf.Sprintf(
			scipMigratorScanResultChunksQuery,
			uploadID,
			pq.Array(ids),
		))); err != nil {
			return err
		}
	}

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

var writers = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

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
) (err error) {
	tr, ctx := trace.New(ctx, "ERICK.HAX.processDocument", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

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

	if err := scipWriter.Write(
		ctx,
		uploadID,
		path,
		scipDocument,
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
) (_ map[int]ResultChunkData, err error) {
	tr, ctx := trace.New(ctx, "ERICK.HAX.fetchResultChunks", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

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
	tx              *basestore.Store
	inserter        *batch.Inserter
	uploadID        int
	nextID          int
	batchPayloadSum int
	batch           []bufferedDocument
}

type bufferedDocument struct {
	path         string
	scipDocument *ogscip.Document
	payload      []byte
	payloadHash  []byte
}

// makeSCIPWriter creates a small wrapper over batch inserts of SCIP data. Each document
// should be written to Postgres by calling Write. The Flush method should be called after
// each document has been processed.
func makeSCIPWriter(ctx context.Context, tx *basestore.Store, uploadID int) (*scipWriter, error) {
	inserter := batch.NewInserter(
		ctx,
		tx.Handle(),
		"codeintel_scip_symbols",
		batch.MaxNumPostgresParameters,
		"upload_id",
		"document_lookup_id",
		"symbol_id",
		"schema_version",
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

// Write inserts a new document and document lookup row, and pushes all of the given
// symbols into the batch inserter.
func (s *scipWriter) Write(
	ctx context.Context,
	uploadID int,
	path string,
	scipDocument *ogscip.Document,
) (err error) {
	tr, ctx := trace.New(ctx, "ERICK.HAX.scipWriter.Write", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	payload, err := proto.Marshal(scipDocument)
	if err != nil {
		return err
	}

	gzipWriter := writers.Get().(*gzip.Writer)
	defer writers.Put(gzipWriter)
	compressBuf := new(bytes.Buffer)
	gzipWriter.Reset(compressBuf)

	if _, err := io.Copy(gzipWriter, bytes.NewReader(payload)); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	uniquePrefix := []byte(fmt.Sprintf(
		"lsif-%d:%d:",
		uploadID,
		time.Now().UnixNano()/int64(time.Millisecond)),
	)

	if s.batchPayloadSum >= MaxBatchPayloadSum {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	s.batch = append(s.batch, bufferedDocument{
		path:         path,
		scipDocument: scipDocument,
		payload:      compressBuf.Bytes(),
		payloadHash:  append(uniquePrefix, hashPayload(payload)...),
	})
	s.batchPayloadSum += len(payload)

	if len(s.batch) >= DocumentsBatchSize {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	return nil
}

// TODO - document
const DocumentsBatchSize = 256
const MaxBatchPayloadSum = 1024 * 1024 * 32

func (s *scipWriter) flush(ctx context.Context) (err error) {
	tr, ctx := trace.New(ctx, "ERICK.HAX.scipWriter.flush", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	documents := s.batch
	s.batch = nil
	s.batchPayloadSum = 0

	documentIDs, err := batchInsertForIdentifiers(
		ctx,
		s.tx.Handle(),
		"codeintel_scip_documents",
		[]string{
			"schema_version",
			"payload_hash",
			"raw_scip_payload",
		},
		func(inserter *batch.Inserter) error {
			for _, document := range documents {
				if err := inserter.Insert(ctx, 2, document.payloadHash, document.payload); err != nil {
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
		panic("OH NO WTF")
	}

	documentLookupIDs, err := batchInsertForIdentifiers(
		ctx,
		s.tx.Handle(),
		"codeintel_scip_document_lookup",
		[]string{
			"upload_id",
			"document_path",
			"document_id",
		},
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
		panic("OH NO WTF")
	}

	symbolCount := 0
	elapsed := time.Duration(0)
	elapsed2 := time.Duration(0)
	elapsed3 := time.Duration(0)

	s1 := time.Now()
	var invertedRangeIndexForDocuments [][]types.InvertedRangeIndex
	for _, document := range documents {
		invertedRangeIndexForDocuments = append(invertedRangeIndexForDocuments, types.ExtractSymbolIndexes(document.scipDocument))
	}

	symbolNameMap := map[string]struct{}{}
	for _, invertedRanges := range invertedRangeIndexForDocuments {
		for _, invertedRange := range invertedRanges {
			symbolNameMap[invertedRange.SymbolName] = struct{}{}
		}
	}
	symbolNames := make([]string, 0, len(symbolNameMap))
	for symbolName := range symbolNameMap {
		symbolNames = append(symbolNames, symbolName)
	}
	sort.Strings(symbolNames)
	elapsed += time.Since(s1)

	s3 := time.Now()
	var trie frozenTrieNode
	trie, s.nextID = constructTrie(symbolNames, s.nextID)
	elapsed2 += time.Since(s3)

	symbolNameByIDs := map[int]string{}
	idsBySymbolName := map[string]int{}

	if err := batch.WithInserter(
		ctx,
		s.tx.Handle(),
		"codeintel_scip_symbol_names",
		batch.MaxNumPostgresParameters,
		[]string{
			"id",
			"upload_id",
			"name_segment",
			"prefix_id",
		},
		func(inserter *batch.Inserter) error {
			s3 := time.Now()

			if err := traverseTrie(trie, func(id int, parentID *int, prefix string) error {
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

				// Do not count sql time against elapsed2
				s4 := time.Now()
				defer func() { elapsed2 -= time.Since(s4) }()

				return inserter.Insert(ctx, id, s.uploadID, prefix, parentID)
			}); err != nil {
				return err
			}

			elapsed2 += time.Since(s3)
			return nil
		},
	); err != nil {
		return err
	}

	for i, ss := range invertedRangeIndexForDocuments {
		documentLookupID := documentLookupIDs[i]

		for _, symbol := range ss {
			symbolCount++

			s2 := time.Now()
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
			elapsed += time.Since(s2)

			s3 := time.Now()
			symbolID, ok := idsBySymbolName[symbol.SymbolName]
			if !ok {
				fmt.Printf("NO SUCH GUY FELLA!\n: %q %v\n", symbol.SymbolName, idsBySymbolName)
				return errors.Newf("failed to construct trie, unknown symbol %q", symbol.SymbolName)
			}
			elapsed3 += time.Since(s3)

			if err := s.inserter.Insert(
				ctx,
				s.uploadID,
				documentLookupID,
				symbolID,
				1,
				definitionRanges,
				referenceRanges,
				implementationRanges,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// Flush ensures that all symbol writes have hit the database, and then moves all of the
// rows from the temporary table into the permanent one.
func (s *scipWriter) Flush(ctx context.Context) (err error) {
	if err := s.flush(ctx); err != nil {
		return err
	}

	if err := s.inserter.Flush(ctx); err != nil {
		return err
	}

	return nil
}

var lsifTableNames = []string{
	"lsif_data_metadata",
	"lsif_data_documents",
	"lsif_data_result_chunks",
	"lsif_data_definitions",
	"lsif_data_references",
	"lsif_data_implementations",
}

func deleteLSIFData(ctx context.Context, tx *basestore.Store, uploadID int) (err error) {
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

// TODO - document
func batchInsertForIdentifiers(
	ctx context.Context,
	db dbutil.DB,
	tableName string,
	columnNames []string,
	f func(inserter *batch.Inserter) error,
) (ids []int, err error) {
	if err := batch.WithInserterWithReturn(
		ctx,
		db,
		tableName,
		batch.MaxNumPostgresParameters,
		columnNames,
		"",
		[]string{"id"},
		func(rows dbutil.Scanner) error {
			id, err := basestore.ScanInt(rows)
			if err != nil {
				return err
			}

			ids = append(ids, id)
			return nil
		},
		f,
	); err != nil {
		return nil, err
	}

	return ids, nil
}

package codeintel

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"google.golang.org/protobuf/proto"
	"k8s.io/utils/lru"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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

func (m *scipMigrator) Up(ctx context.Context) (err error) {
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

const scipMigrationDocumentBatchSize = 128
const scipMigratorResultChunkDefaultCacheSize = 1024

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
			scipMigrationDocumentBatchSize,
			page*scipMigrationDocumentBatchSize,
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

	scipDocument := convertDocument(
		uploadID,
		resultChunks,
		numResultChunks,
		indexerName,
		path,
		document,
	)

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
		idx := hashKey(id, numResultChunks)

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

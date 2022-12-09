package lsifstore

import (
	"context"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type LsifStore interface {
	Transact(ctx context.Context) (LsifStore, error)
	Done(err error) error

	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	DeleteLsifDataByUploadIds(ctx context.Context, bundleIDs ...int) (err error)

	InsertMetadata(ctx context.Context, uploadID int, meta ProcessedMetadata) error
	NewSymbolWriter(ctx context.Context, uploadID int) (SymbolWriter, error)
	InsertSCIPDocument(ctx context.Context, uploadID int, documentPath string, hash []byte, rawSCIPPayload []byte) (int, error)

	WriteMeta(ctx context.Context, bundleID int, meta precise.MetaData) error
	WriteDocuments(ctx context.Context, bundleID int, documents chan precise.KeyedDocumentData) (count uint32, err error)
	WriteResultChunks(ctx context.Context, bundleID int, resultChunks chan precise.IndexedResultChunkData) (count uint32, err error)
	WriteDefinitions(ctx context.Context, bundleID int, monikerLocations chan precise.MonikerLocations) (count uint32, err error)
	WriteReferences(ctx context.Context, bundleID int, monikerLocations chan precise.MonikerLocations) (count uint32, err error)
	WriteImplementations(ctx context.Context, bundleID int, monikerLocations chan precise.MonikerLocations) (count uint32, err error)

	IDsWithMeta(ctx context.Context, ids []int) ([]int, error)
	ReconcileCandidates(ctx context.Context, batchSize int) ([]int, error)

	// Stream
	ScanDocuments(ctx context.Context, id int, f func(path string, ranges map[precise.ID]precise.RangeData) error) (err error)
	ScanResultChunks(ctx context.Context, id int, f func(idx int, resultChunk precise.ResultChunkData) error) (err error)
	ScanLocations(ctx context.Context, id int, f func(scheme, identifier, monikerType string, locations []precise.LocationData) error) (err error)
}

type SymbolWriter interface {
	WriteSCIPSymbols(ctx context.Context, documentLookupID int, symbols []types.InvertedRangeIndex) error
	Flush(ctx context.Context) (uint32, error)
}

type store struct {
	db         *basestore.Store
	serializer *Serializer
	operations *operations
}

func New(observationCtx *observation.Context, db codeintelshared.CodeIntelDB) LsifStore {
	return newStore(observationCtx, db)
}

func newStore(observationCtx *observation.Context, db codeintelshared.CodeIntelDB) *store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		serializer: NewSerializer(),
		operations: newOperations(observationCtx),
	}
}

func (s *store) Transact(ctx context.Context) (LsifStore, error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		db:         tx,
		serializer: s.serializer,
		operations: s.operations,
	}, nil
}

func (s *store) Done(err error) error {
	return s.db.Done(err)
}

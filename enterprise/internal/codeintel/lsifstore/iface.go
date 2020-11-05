package lsifstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

// KeyedDocumentData pairs a document with its path.
type KeyedDocumentData struct {
	Path     string
	Document types.DocumentData
}

// IndexedResultChunkData pairs a result chunk with its index.
type IndexedResultChunkData struct {
	Index       int
	ResultChunk types.ResultChunkData
}

type Store interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	Clear(ctx context.Context, bundleIDs ...int) error

	ReadMeta(ctx context.Context, bundleID int) (types.MetaData, error)
	PathsWithPrefix(ctx context.Context, bundleID int, prefix string) ([]string, error)
	ReadDocument(ctx context.Context, bundleID int, path string) (types.DocumentData, bool, error)
	ReadResultChunk(ctx context.Context, bundleID int, id int) (types.ResultChunkData, bool, error)
	ReadDefinitions(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) ([]types.Location, int, error)
	ReadReferences(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) ([]types.Location, int, error)

	WriteMeta(ctx context.Context, bundleID int, meta types.MetaData) error
	WriteDocuments(ctx context.Context, bundleID int, documents chan KeyedDocumentData) error
	WriteResultChunks(ctx context.Context, bundleID int, resultChunks chan IndexedResultChunkData) error
	WriteDefinitions(ctx context.Context, bundleID int, monikerLocations chan types.MonikerLocations) error
	WriteReferences(ctx context.Context, bundleID int, monikerLocations chan types.MonikerLocations) error
}

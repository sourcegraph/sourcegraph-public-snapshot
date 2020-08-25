package persistence

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

type Writer interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error
	CreateTables(ctx context.Context) error

	WriteMeta(ctx context.Context, meta types.MetaData) error
	WriteDocuments(ctx context.Context, documents chan KeyedDocumentData) error
	WriteResultChunks(ctx context.Context, resultChunks chan IndexedResultChunkData) error
	WriteDefinitions(ctx context.Context, monikerLocations chan types.MonikerLocations) error
	WriteReferences(ctx context.Context, monikerLocations chan types.MonikerLocations) error
	Close(err error) error
}

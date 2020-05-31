package persistence

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type Writer interface {
	WriteMeta(ctx context.Context, metaData types.MetaData) error
	WriteDocuments(ctx context.Context, documents map[string]types.DocumentData) error
	WriteResultChunks(ctx context.Context, resultChunks map[int]types.ResultChunkData) error
	WriteDefinitions(ctx context.Context, monikerLocations []types.MonikerLocations) error
	WriteReferences(ctx context.Context, monikerLocations []types.MonikerLocations) error
	Flush(ctx context.Context) error
	Close() error
}

package persistence

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type Reader interface {
	ReadMeta(ctx context.Context) (types.MetaData, error)
	ReadDocument(ctx context.Context, path string) (types.DocumentData, bool, error)
	ReadResultChunk(ctx context.Context, id int) (types.ResultChunkData, bool, error)
	ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error)
	ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error)
	Close() error
}

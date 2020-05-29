package persistence

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type Writer interface {
	WriteMeta(ctx context.Context, lsifVersion string, numResultChunks int) error
	WriteDocuments(ctx context.Context, documents map[string]types.DocumentData) error
	WriteResultChunks(ctx context.Context, resultChunks map[int]types.ResultChunkData) error
	WriteDefinitions(ctx context.Context, definitions []types.DefinitionReferenceRow) error
	WriteReferences(ctx context.Context, references []types.DefinitionReferenceRow) error
	Flush(ctx context.Context) error
	Close() error
}

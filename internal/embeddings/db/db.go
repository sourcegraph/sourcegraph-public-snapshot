package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type VectorDB interface {
	VectorSearcher
	VectorInserter
}

type VectorSearcher interface {
	Search(context.Context, SearchParams) ([]ChunkResult, error)
}

type VectorInserter interface {
	PrepareUpdate(ctx context.Context, modelID string, modelDims uint64) error
	HasIndex(ctx context.Context, modelID string, repoID api.RepoID, revision api.CommitID) (bool, error)
	InsertChunks(context.Context, InsertParams) error
	FinalizeUpdate(context.Context, FinalizeUpdateParams) error
}

func CollectionName(modelID string) string {
	return fmt.Sprintf("repos.%s", strings.ReplaceAll(modelID, "/", "."))
}

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

type SearchParams struct {
	// RepoIDs is the set of repos to search.
	// If empty, all repos are searched.
	RepoIDs []api.RepoID

	// The ID of the model that the query was embedded with.
	// Embeddings for other models will not be searched.
	ModelID string

	// Query is the embedding for the search query.
	// Its dimensions must match the model dimensions.
	Query []float32

	// The maximum number of code results to return
	CodeLimit int

	// The maximum number of text results to return
	TextLimit int
}

type InsertParams struct {
	ModelID     string
	ChunkPoints ChunkPoints
}

type FinalizeUpdateParams struct {
	ModelID       string
	RepoID        api.RepoID
	Revision      api.CommitID
	FilesToRemove []string
}

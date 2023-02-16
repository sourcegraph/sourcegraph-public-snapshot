package shared

import (
	"context"
	"encoding/json"
	"io"

	lru "github.com/hashicorp/golang-lru"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

func downloadRepoEmbeddingIndex(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName, uploadStore uploadstore.Store) (*embeddings.RepoEmbeddingIndex, error) {
	repoEmbeddingIndexFile, err := uploadStore.Get(ctx, string(repoEmbeddingIndexName))
	if err != nil {
		return nil, err
	}

	repoEmbeddingIndexFileBytes, err := io.ReadAll(repoEmbeddingIndexFile)
	if err != nil {
		return nil, err
	}

	var embeddingIndex embeddings.RepoEmbeddingIndex
	err = json.Unmarshal(repoEmbeddingIndexFileBytes, &embeddingIndex)
	if err != nil {
		return nil, err
	}
	return &embeddingIndex, nil
}

func getCachedRepoEmbeddingIndex(ctx context.Context, cache *lru.Cache, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName, uploadStore uploadstore.Store) (*embeddings.RepoEmbeddingIndex, error) {
	var err error
	var embeddingIndex *embeddings.RepoEmbeddingIndex

	if cachedEmbeddingIndex, ok := cache.Get(repoEmbeddingIndexName); ok {
		embeddingIndex = cachedEmbeddingIndex.(*embeddings.RepoEmbeddingIndex)
	} else {
		embeddingIndex, err = downloadRepoEmbeddingIndex(ctx, repoEmbeddingIndexName, uploadStore)
		if err != nil {
			return nil, err
		}
		cache.Add(repoEmbeddingIndexName, embeddingIndex)
	}

	return embeddingIndex, err
}

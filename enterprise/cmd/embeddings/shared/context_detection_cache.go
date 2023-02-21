package shared

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

func downloadContextDetectionEmbeddingIndex(ctx context.Context, uploadStore uploadstore.Store) (*embeddings.ContextDetectionEmbeddingIndex, error) {
	indexFile, err := uploadStore.Get(ctx, embeddings.CONTEXT_DETECTION_INDEX_NAME)
	if err != nil {
		return nil, err
	}

	indexFileBytes, err := io.ReadAll(indexFile)
	if err != nil {
		return nil, err
	}

	var embeddingIndex embeddings.ContextDetectionEmbeddingIndex
	err = json.Unmarshal(indexFileBytes, &embeddingIndex)
	if err != nil {
		return nil, err
	}
	return &embeddingIndex, nil
}

func getCachedContextDetectionEmbeddingIndexFn(uploadStore uploadstore.Store) getContextDetectionEmbeddingIndexFn {
	var contextDetectionEmbeddingIndex *embeddings.ContextDetectionEmbeddingIndex = nil

	return func(ctx context.Context) (*embeddings.ContextDetectionEmbeddingIndex, error) {
		if contextDetectionEmbeddingIndex != nil {
			return contextDetectionEmbeddingIndex, nil
		}
		var err error
		contextDetectionEmbeddingIndex, err = downloadContextDetectionEmbeddingIndex(ctx, uploadStore)
		if err != nil {
			return nil, err
		}
		return contextDetectionEmbeddingIndex, nil
	}
}

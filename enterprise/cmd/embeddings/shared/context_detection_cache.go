package shared

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

func getCachedContextDetectionEmbeddingIndex(uploadStore uploadstore.Store) getContextDetectionEmbeddingIndexFn {
	mu := sync.Mutex{}
	var contextDetectionEmbeddingIndex *embeddings.ContextDetectionEmbeddingIndex = nil
	return func(ctx context.Context) (_ *embeddings.ContextDetectionEmbeddingIndex, err error) {
		mu.Lock()
		defer mu.Unlock()
		if contextDetectionEmbeddingIndex != nil {
			return contextDetectionEmbeddingIndex, nil
		}
		contextDetectionEmbeddingIndex, err = embeddings.DownloadIndex[embeddings.ContextDetectionEmbeddingIndex](ctx, uploadStore, embeddings.CONTEXT_DETECTION_INDEX_NAME)
		if err != nil {
			return nil, err
		}
		return contextDetectionEmbeddingIndex, nil
	}
}

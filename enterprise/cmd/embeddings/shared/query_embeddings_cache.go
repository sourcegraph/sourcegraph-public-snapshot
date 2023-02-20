package shared

import (
	lru "github.com/hashicorp/golang-lru"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/schema"
)

func getCachedQueryEmbedding(cache *lru.Cache, config *schema.Embeddings, query string) ([]float32, error) {
	var err error
	var queryEmbedding []float32

	if cachedQueryEmbedding, ok := cache.Get(query); ok {
		queryEmbedding = cachedQueryEmbedding.([]float32)
	} else {
		queryEmbedding, err = embed.GetEmbeddingsWithRetries([]string{query}, config, QUERY_EMBEDDING_RETRIES)
		if err != nil {
			return nil, err
		}
		cache.Add(query, queryEmbedding)
	}
	return queryEmbedding, err
}

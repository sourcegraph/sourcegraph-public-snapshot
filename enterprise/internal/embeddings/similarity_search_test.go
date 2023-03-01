package embeddings

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSimilaritySearch(t *testing.T) {
	embeddings := []float32{
		0.26726124, 0.53452248, 0.80178373,
		0.45584231, 0.56980288, 0.68376346,
		0.50257071, 0.57436653, 0.64616234,
	}
	index := EmbeddingIndex[RepoEmbeddingRowMetadata]{
		Embeddings:      embeddings,
		ColumnDimension: 3,
		RowMetadata: []RepoEmbeddingRowMetadata{
			{FileName: "a"},
			{FileName: "b"},
			{FileName: "c"},
		},
	}

	t.Run("find row with exact match", func(t *testing.T) {
		query := embeddings[0:3]
		results := index.SimilaritySearch(query, 1)
		autogold.Equal(t, results)
	})

	t.Run("find nearest neighbors", func(t *testing.T) {
		query := []float32{0.87006284, 0.48336824, 0.09667365}
		results := index.SimilaritySearch(query, 2)
		autogold.Equal(t, results)
	})

	t.Run("request more results then there are rows", func(t *testing.T) {
		query := embeddings[0:3]
		results := index.SimilaritySearch(query, 5)
		autogold.Equal(t, results)
	})
}

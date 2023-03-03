package embeddings

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// Each line represents a separate embedding.
var embeddings = []float32{
	0.5061, 0.6595, 0.5558,
	0.5764, 0.4482, 0.6833,
	0.3162, 0.6444, 0.6963,
	0.3736, 0.7713, 0.5153,
	0.5219, 0.8505, 0.0653,
	0.1040, 0.0241, 0.9943,
	0.5109, 0.5712, 0.6425,
	0.6612, 0.3818, 0.6458,
	0.1775, 0.9603, 0.2151,
	0.8171, 0.4514, 0.3587,
	0.2824, 0.8265, 0.4869,
	0.6770, 0.0224, 0.7356,
	0.4771, 0.4809, 0.7356,
	0.7695, 0.4057, 0.4932,
	0.7215, 0.0623, 0.6896,
	0.9385, 0.2944, 0.1804,
}

// Each line represents a separate query.
var queries = []float32{
	0.4227, 0.4874, 0.7641,
	0.4038, 0.9100, 0.0940,
	0.2965, 0.2290, 0.9272,
}

// Each subarray contains ranked nearest neighbors for each query.
var ranks = [][]int{
	{4, 2, 3, 6, 14, 12, 1, 5, 13, 11, 8, 10, 0, 7, 9, 15},
	{4, 9, 6, 3, 0, 15, 5, 11, 1, 7, 2, 14, 10, 8, 13, 12},
	{8, 2, 4, 10, 15, 0, 6, 5, 14, 12, 11, 3, 1, 9, 7, 13},
}

func TestSimilaritySearch(t *testing.T) {
	nRows, nQueries, columnDimension := 16, 3, 3
	index := EmbeddingIndex[RepoEmbeddingRowMetadata]{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		RowMetadata:     []RepoEmbeddingRowMetadata{},
	}

	for i := 0; i < nRows; i++ {
		index.RowMetadata = append(index.RowMetadata, RepoEmbeddingRowMetadata{FileName: fmt.Sprintf("%d", i)})
	}

	getExpectedResults := func(queryRanks []int) []*RepoEmbeddingRowMetadata {
		results := make([]*RepoEmbeddingRowMetadata, len(queryRanks))
		for idx, rank := range queryRanks {
			results[rank] = &index.RowMetadata[idx]
		}
		return results
	}

	for _, nWorkers := range []int{0, 1, 2, 3, 5, 8, 9, 16, 20, 33} {
		for _, nResults := range []int{0, 1, 2, 4, 9, 16, 32} {
			for q := 0; q < nQueries; q++ {
				t.Run(fmt.Sprintf("find nearest neighbors, query=%d, nResults=%d, nWorkers=%d", q, nResults, nWorkers), func(t *testing.T) {
					query := queries[q*columnDimension : (q+1)*columnDimension]
					results := index.SimilaritySearch(query, nResults, nWorkers)
					expectedResults := getExpectedResults(ranks[q])
					require.Equal(t, expectedResults[:min(nResults, len(expectedResults))], results)
				})
			}
		}
	}
}

func TestSplitRows(t *testing.T) {
	tests := []struct {
		nRows    int
		nWorkers int
		want     []partialRows
	}{
		{
			nRows:    0,
			nWorkers: 1,
			want:     []partialRows{{0, 0}},
		},
		{
			nRows:    128,
			nWorkers: 1,
			want:     []partialRows{{0, 128}},
		},
		{
			nRows:    16,
			nWorkers: 4,
			want:     []partialRows{{0, 4}, {4, 8}, {8, 12}, {12, 16}},
		},
		{
			nRows:    5,
			nWorkers: 4,
			want:     []partialRows{{0, 2}, {2, 4}, {4, 5}, {5, 5}},
		},
		{
			nRows:    16,
			nWorkers: 3,
			want:     []partialRows{{0, 6}, {6, 12}, {12, 16}},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("nRows=%d, nWorkers=%d", tt.nRows, tt.nWorkers), func(t *testing.T) {
			got := splitRows(tt.nRows, tt.nWorkers)
			require.Equal(t, tt.want, got)
		})
	}
}

func getRandomEmbeddings(prng *rand.Rand, nElements int) []float32 {
	slice := make([]float32, nElements)
	for idx := range slice {
		slice[idx] = prng.Float32()
	}
	return slice
}

func BenchmarkSimilaritySearch(b *testing.B) {
	prng := rand.New(rand.NewSource(0))

	nRows := 1_000_000
	nResults := 100
	columnDimension := 1536
	index := &EmbeddingIndex[RepoEmbeddingRowMetadata]{
		Embeddings:      getRandomEmbeddings(prng, nRows*columnDimension),
		ColumnDimension: columnDimension,
		RowMetadata:     make([]RepoEmbeddingRowMetadata, nRows),
	}
	query := getRandomEmbeddings(prng, columnDimension)

	b.ResetTimer()

	for _, nWorkers := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("nWorkers=%d", nWorkers), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = index.SimilaritySearch(query, nResults, nWorkers)
			}
		})
	}
}

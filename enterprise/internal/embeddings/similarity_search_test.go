package embeddings

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// Data below was generated using the testdata/generate_similarity_search_test_data.py script.

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
	numRows, numQueries, columnDimension := 16, 3, 3
	index := EmbeddingIndex[RepoEmbeddingRowMetadata]{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		RowMetadata:     []RepoEmbeddingRowMetadata{},
	}

	for i := 0; i < numRows; i++ {
		index.RowMetadata = append(index.RowMetadata, RepoEmbeddingRowMetadata{FileName: fmt.Sprintf("%d", i)})
	}

	getExpectedResults := func(queryRanks []int) []*RepoEmbeddingRowMetadata {
		results := make([]*RepoEmbeddingRowMetadata, len(queryRanks))
		for idx, rank := range queryRanks {
			results[rank] = &index.RowMetadata[idx]
		}
		return results
	}

	for _, numWorkers := range []int{0, 1, 2, 3, 5, 8, 9, 16, 20, 33} {
		for _, numResults := range []int{0, 1, 2, 4, 9, 16, 32} {
			for q := 0; q < numQueries; q++ {
				t.Run(fmt.Sprintf("find nearest neighbors query=%d numResults=%d numWorkers=%d", q, numResults, numWorkers), func(t *testing.T) {
					query := queries[q*columnDimension : (q+1)*columnDimension]
					results := index.SimilaritySearch(query, numResults, WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: 0})
					expectedResults := getExpectedResults(ranks[q])
					require.Equal(t, expectedResults[:min(numResults, len(expectedResults))], results)
				})
			}
		}
	}
}

func TestSplitRows(t *testing.T) {
	tests := []struct {
		numRows        int
		numWorkers     int
		minRowsToSplit int
		want           []partialRows
	}{
		{
			numRows:    0,
			numWorkers: 1,
			want:       []partialRows{{0, 0}},
		},
		{
			numRows:    128,
			numWorkers: 1,
			want:       []partialRows{{0, 128}},
		},
		{
			numRows:    16,
			numWorkers: 4,
			want:       []partialRows{{0, 4}, {4, 8}, {8, 12}, {12, 16}},
		},
		{
			numRows:    5,
			numWorkers: 4,
			want:       []partialRows{{0, 2}, {2, 4}, {4, 5}, {5, 5}},
		},
		{
			numRows:    16,
			numWorkers: 3,
			want:       []partialRows{{0, 6}, {6, 12}, {12, 16}},
		},
		{
			numRows:        20,
			numWorkers:     5,
			minRowsToSplit: 20,
			want:           []partialRows{{0, 20}},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("numRows=%d numWorkers=%d", tt.numRows, tt.numWorkers), func(t *testing.T) {
			got := splitRows(tt.numRows, tt.numWorkers, tt.minRowsToSplit)
			require.Equal(t, tt.want, got)
		})
	}
}

func getRandomEmbeddings(prng *rand.Rand, numElements int) []float32 {
	slice := make([]float32, numElements)
	for idx := range slice {
		slice[idx] = prng.Float32()
	}
	return slice
}

func BenchmarkSimilaritySearch(b *testing.B) {
	prng := rand.New(rand.NewSource(0))

	numRows := 1_000_000
	numResults := 100
	columnDimension := 1536
	index := &EmbeddingIndex[RepoEmbeddingRowMetadata]{
		Embeddings:      getRandomEmbeddings(prng, numRows*columnDimension),
		ColumnDimension: columnDimension,
		RowMetadata:     make([]RepoEmbeddingRowMetadata, numRows),
	}
	query := getRandomEmbeddings(prng, columnDimension)

	b.ResetTimer()

	for _, numWorkers := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("numWorkers=%d", numWorkers), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = index.SimilaritySearch(query, numResults, WorkerOptions{NumWorkers: numWorkers})
			}
		})
	}
}

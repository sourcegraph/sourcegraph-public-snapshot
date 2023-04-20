package embeddings

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// Data below was generated using the testdata/generate_similarity_search_test_data.py script.

// Each line represents a separate embedding.
var embeddings = []int8{
	64, 83, 70,
	73, 56, 86,
	40, 81, 88,
	47, 97, 65,
	66, 108, 8,
	13, 3, 126,
	64, 72, 81,
	83, 48, 82,
	22, 121, 27,
	103, 57, 45,
	35, 104, 61,
	85, 2, 93,
	60, 61, 93,
	97, 51, 62,
	91, 7, 87,
	119, 37, 22,
}

// Each line represents a separate query.
var queries = []int8{
	53, 61, 97,
	51, 115, 11,
	37, 29, 117,
}

// Each subarray contains ranked nearest neighbors for each query.
var ranks = [][]int{
	{4, 2, 3, 6, 14, 12, 1, 5, 13, 11, 8, 10, 0, 7, 9, 15},
	{4, 9, 6, 3, 0, 15, 5, 11, 1, 7, 2, 14, 10, 8, 13, 12},
	{8, 2, 4, 10, 15, 0, 6, 5, 14, 12, 11, 3, 1, 9, 7, 13},
}

func TestSimilaritySearch(t *testing.T) {
	numRows, numQueries, columnDimension := 16, 3, 3
	index := EmbeddingIndex{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		RowMetadata:     []RepoEmbeddingRowMetadata{},
	}

	for i := 0; i < numRows; i++ {
		index.RowMetadata = append(index.RowMetadata, RepoEmbeddingRowMetadata{FileName: fmt.Sprintf("%d", i)})
	}

	getExpectedResults := func(queryRanks []int) []EmbeddingSearchResult {
		results := make([]EmbeddingSearchResult, len(queryRanks))
		for idx, rank := range queryRanks {
			results[rank].RepoEmbeddingRowMetadata = index.RowMetadata[idx]
		}
		return results
	}

	for _, numWorkers := range []int{0, 1, 2, 3, 5, 8, 9, 16, 20, 33} {
		for _, numResults := range []int{0, 1, 2, 4, 9, 16, 32} {
			for q := 0; q < numQueries; q++ {
				t.Run(fmt.Sprintf("find nearest neighbors query=%d numResults=%d numWorkers=%d", q, numResults, numWorkers), func(t *testing.T) {
					query := queries[q*columnDimension : (q+1)*columnDimension]
					results := index.SimilaritySearch(query, numResults, WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: 0}, SearchOptions{})
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

func getRandomEmbeddings(prng *rand.Rand, numElements int) []int8 {
	slice := make([]int8, numElements)
	for idx := range slice {
		slice[idx] = int8(prng.Int())
	}
	return slice
}

func BenchmarkSimilaritySearch(b *testing.B) {
	prng := rand.New(rand.NewSource(0))

	numRows := 1_000_000
	numResults := 100
	columnDimension := 1536
	index := &EmbeddingIndex{
		Embeddings:      getRandomEmbeddings(prng, numRows*columnDimension),
		ColumnDimension: columnDimension,
		RowMetadata:     make([]RepoEmbeddingRowMetadata, numRows),
	}
	query := getRandomEmbeddings(prng, columnDimension)

	b.ResetTimer()

	for _, numWorkers := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("numWorkers=%d", numWorkers), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = index.SimilaritySearch(query, numResults, WorkerOptions{NumWorkers: numWorkers}, SearchOptions{})
			}
		})
	}
}

func TestScore(t *testing.T) {
	var ranks []float32
	for i := 1; i < len(embeddings); i++ {
		ranks = append(ranks, float32(i))
	}

	columnDimension := 3
	index := &EmbeddingIndex{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		Ranks:           ranks,
	}
	// embeddings[0] = 64, 83, 70,
	// queries[0:3] = 53, 61, 97,
	score, debugInfo := index.score(queries[0:columnDimension], 0, SearchOptions{Debug: true, UseDocumentRanks: true})

	// Check that the score is correct
	expectedScore := scoreSimilarityWeight * ((64 * 53) + (83 * 61) + (70 * 97))
	if math.Abs(float64(score-expectedScore)) > 0.0001 {
		t.Fatalf("Expected score %d, but got %d", expectedScore, score)
	}

	if debugInfo == "" {
		t.Fatal("Expected a non-empty debug string")
	}
}

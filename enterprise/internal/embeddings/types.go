package embeddings

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type EmbeddingIndex struct {
	Embeddings      []int8
	ColumnDimension int
	RowMetadata     []RepoEmbeddingRowMetadata
	Ranks           []float32
}

// Row returns the embeddings for the nth row in the index
func (index *EmbeddingIndex) Row(n int) []int8 {
	return index.Embeddings[n*index.ColumnDimension : (n+1)*index.ColumnDimension]
}

type RepoEmbeddingRowMetadata struct {
	FileName  string `json:"fileName"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

type RepoEmbeddingIndex struct {
	RepoName  api.RepoName
	Revision  api.CommitID
	CodeIndex EmbeddingIndex
	TextIndex EmbeddingIndex
}

type ContextDetectionEmbeddingIndex struct {
	MessagesWithAdditionalContextMeanEmbedding    []float32
	MessagesWithoutAdditionalContextMeanEmbedding []float32
}

type EmbeddingSearchResults struct {
	CodeResults []EmbeddingSearchResult `json:"codeResults"`
	TextResults []EmbeddingSearchResult `json:"textResults"`
}

type EmbeddingSearchResult struct {
	RepoEmbeddingRowMetadata
	// The row number in the index to correlate this result back with its source.
	RowNum  int    `json:"rowNum"`
	Content string `json:"content"`
	// Experimental: Clients should not rely on any particular format of debug
	Debug string `json:"debug,omitempty"`
}

type EmbedRepoStats struct {
	// Repo name
	RepoName api.RepoName
	Revision api.CommitID

	HasRanks       bool
	InputFileCount int

	CodeIndexStats EmbedFilesStats
	TextIndexStats EmbedFilesStats
}

type EmbedFilesStats struct {
	// The time it took to generate these embeddings
	Duration time.Duration

	// The number of files embedded
	EmbeddedCount int

	// The sum of the size of the contents of successful embeddings
	EmbeddedBytes int

	// Summed byte counts for each of the reasons files were skipped
	SkippedByteCounts map[string]int

	// Counts of reasons files were skipped
	SkippedCounts map[string]int
}

package embeddings

import (
	"time"

	"github.com/sourcegraph/log"
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

// DEPRECATED: to support decoding old indexes, we need a struct
// we can decode into directly. This struct is the same shape
// as the old indexes and should not be changed without migrating
// all existing indexes to the new format.
type OldRepoEmbeddingIndex struct {
	RepoName  api.RepoName
	Revision  api.CommitID
	CodeIndex OldEmbeddingIndex
	TextIndex OldEmbeddingIndex
}

func (o *OldRepoEmbeddingIndex) ToNewIndex() *RepoEmbeddingIndex {
	return &RepoEmbeddingIndex{
		RepoName:  o.RepoName,
		Revision:  o.Revision,
		CodeIndex: o.CodeIndex.ToNewIndex(),
		TextIndex: o.TextIndex.ToNewIndex(),
	}
}

type OldEmbeddingIndex struct {
	Embeddings      []float32
	ColumnDimension int
	RowMetadata     []RepoEmbeddingRowMetadata
	Ranks           []float32
}

func (o *OldEmbeddingIndex) ToNewIndex() EmbeddingIndex {
	return EmbeddingIndex{
		Embeddings:      Quantize(o.Embeddings),
		ColumnDimension: o.ColumnDimension,
		RowMetadata:     o.RowMetadata,
		Ranks:           o.Ranks,
	}
}

type EmbedRepoStats struct {
	Duration       time.Duration
	HasRanks       bool
	CodeIndexStats EmbedFilesStats
	TextIndexStats EmbedFilesStats
}

func (e *EmbedRepoStats) ToFields() []log.Field {
	return []log.Field{
		log.Duration("duration", e.Duration),
		log.Bool("hasRanks", e.HasRanks),
		log.Object("codeIndex", e.CodeIndexStats.ToFields()...),
		log.Object("textIndex", e.TextIndexStats.ToFields()...),
	}
}

type EmbedFilesStats struct {
	// The time it took to generate these embeddings
	Duration time.Duration

	// The number of files embedded
	EmbeddedFileCount int

	// The number of chunks we generated embeddings for.
	// Equivalent to the number of embeddings generated.
	EmbeddedChunkCount int

	// The sum of the size of the contents of successful embeddings
	EmbeddedBytes int

	// Summed byte counts for each of the reasons files were skipped
	SkippedByteCounts map[string]int

	// Counts of reasons files were skipped
	SkippedCounts map[string]int
}

func (e *EmbedFilesStats) ToFields() []log.Field {
	var skippedByteCounts []log.Field
	for reason, count := range e.SkippedByteCounts {
		skippedByteCounts = append(skippedByteCounts, log.Int(reason, count))
	}

	var skippedCounts []log.Field
	for reason, count := range e.SkippedCounts {
		skippedCounts = append(skippedCounts, log.Int(reason, count))
	}
	return []log.Field{
		log.Duration("duration", e.Duration),
		log.Int("embeddedFileCount", e.EmbeddedFileCount),
		log.Int("embeddedChunkCount", e.EmbeddedChunkCount),
		log.Int("embeddedBytes", e.EmbeddedBytes),
		log.Object("skippedCounts", skippedCounts...),
		log.Object("skippedByteCounts", skippedByteCounts...),
	}
}

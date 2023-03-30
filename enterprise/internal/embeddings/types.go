package embeddings

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type EmbeddingIndex struct {
	Embeddings      []float32
	ColumnDimension int
	RowMetadata     []RepoEmbeddingRowMetadata
	Ranks           []float32
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
	Content string `json:"content"`
	// Experimental: Clients should not rely on any particular format of debug
	Debug string `json:"debug,omitempty"`
}

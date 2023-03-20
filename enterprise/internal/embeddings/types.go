package embeddings

import "github.com/sourcegraph/sourcegraph/internal/api"

type EmbeddingIndex[T any] struct {
	Embeddings      []float32
	ColumnDimension int
	RowMetadata     []T
	Ranks           []float64
}

type RepoEmbeddingRowMetadata struct {
	FileName  string
	StartLine int
	EndLine   int
}

type RepoEmbeddingIndex struct {
	RepoName  api.RepoName
	Revision  api.CommitID
	CodeIndex EmbeddingIndex[RepoEmbeddingRowMetadata]
	TextIndex EmbeddingIndex[RepoEmbeddingRowMetadata]
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
	FileName  string `json:"fileName"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Content   string `json:"content"`
}

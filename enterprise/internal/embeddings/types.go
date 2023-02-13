package embeddings

import "github.com/sourcegraph/sourcegraph/internal/api"

type RepoEmbeddingIndex struct {
	RepoName  api.RepoName   `json:"repoName"`
	Revision  api.CommitID   `json:"revision"`
	CodeIndex EmbeddingIndex `json:"code"`
	TextIndex EmbeddingIndex `json:"text"`
}

type EmbeddingIndex struct {
	Embeddings      []float32              `json:"embeddings"`
	ColumnDimension int                    `json:"columnDimension"`
	RowMetadata     []EmbeddingRowMetadata `json:"rowMetadata"`
}

type EmbeddingRowMetadata struct {
	FileName  string `json:"fileName"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

type EmbeddingSearchResults struct {
	CodeResults []EmbeddingSearchResult `json:"codeResults"`
	TextResults []EmbeddingSearchResult `json:"textResults"`
}

type EmbeddingSearchResult struct {
	FilePath  string `json:"filePath"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Content   string `json:"content"`
}

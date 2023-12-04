package embeddings

import (
	"fmt"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (index *EmbeddingIndex) EstimateSize() uint64 {
	return uint64(len(index.Embeddings) + len(index.RowMetadata)*(16+8+8) + len(index.Ranks)*4)
}

// Validate will return a non-nil error if the fields on index break an
// invariant. This is useful to call after unmarshalling to ensure there is no
// corruption.
func (index *EmbeddingIndex) Validate() error {
	if len(index.Embeddings) != index.ColumnDimension*len(index.RowMetadata) {
		return errors.Errorf("embedding index has an unexpected number of cells: cells=%d != columns=%d * rows=%d", len(index.Embeddings), index.ColumnDimension, len(index.RowMetadata))
	}

	return nil
}

// Filter removes all files from the index that are in the set and updates the ranks
func (index *EmbeddingIndex) filter(set map[string]struct{}, ranks types.RepoPathRanks) {
	// We can reset Ranks here because we are anyway going to update them based on
	// "ranks".
	index.Ranks = make([]float32, 0, len(index.RowMetadata))

	cursor := 0
	for i, s := range index.RowMetadata {
		if _, ok := set[s.FileName]; ok {
			continue
		}
		index.RowMetadata[cursor] = s

		// Ranks might have changed since the index was created, so we need to update
		// them
		index.Ranks = append(index.Ranks, float32(ranks.Paths[s.FileName]))

		copy(index.Row(cursor), index.Row(i))
		cursor++
	}

	// update slice length
	index.RowMetadata = index.RowMetadata[:cursor]
	index.Ranks = index.Ranks[:cursor]
	index.Embeddings = index.Embeddings[:cursor*index.ColumnDimension]
}

func (index *EmbeddingIndex) append(other EmbeddingIndex) {
	index.RowMetadata = append(index.RowMetadata, other.RowMetadata...)
	index.Ranks = append(index.Ranks, other.Ranks...)
	index.Embeddings = append(index.Embeddings, other.Embeddings...)
}

type RepoEmbeddingRowMetadata struct {
	FileName  string `json:"fileName"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

type RepoEmbeddingIndex struct {
	RepoName        api.RepoName
	Revision        api.CommitID
	EmbeddingsModel string
	CodeIndex       EmbeddingIndex
	TextIndex       EmbeddingIndex
}

func (i *RepoEmbeddingIndex) EstimateSize() uint64 {
	return i.CodeIndex.EstimateSize() + i.TextIndex.EstimateSize()
}

func (i *RepoEmbeddingIndex) IsModelCompatible(model string) bool {
	return i.EmbeddingsModel == "" || i.EmbeddingsModel == model
}

type ContextDetectionEmbeddingIndex struct {
	MessagesWithAdditionalContextMeanEmbedding    []float32
	MessagesWithoutAdditionalContextMeanEmbedding []float32
}

type EmbeddingCombinedSearchResults struct {
	CodeResults EmbeddingSearchResults `json:"codeResults"`
	TextResults EmbeddingSearchResults `json:"textResults"`
}

type EmbeddingSearchResults []EmbeddingSearchResult

// MergeTruncate merges other into the search results, keeping only max results with the highest scores
func (esrs *EmbeddingSearchResults) MergeTruncate(other EmbeddingSearchResults, max int) {
	self := *esrs
	self = append(self, other...)
	sort.Slice(self, func(i, j int) bool { return self[i].Score() > self[j].Score() })
	if len(self) > max {
		self = self[:max]
	}
	*esrs = self
}

type EmbeddingSearchResult struct {
	RepoName api.RepoName `json:"repoName"`
	Revision api.CommitID `json:"revision"`

	FileName  string `json:"fileName"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`

	ScoreDetails SearchScoreDetails `json:"scoreDetails"`
}

func (esr *EmbeddingSearchResult) Score() int32 {
	return esr.ScoreDetails.RankScore + esr.ScoreDetails.SimilarityScore
}

type SearchScoreDetails struct {
	Score int32 `json:"score"`

	// Breakdown
	SimilarityScore int32 `json:"similarityScore"`
	RankScore       int32 `json:"rankScore"`
}

func (s *SearchScoreDetails) String() string {
	return fmt.Sprintf("score:%d, similarity:%d, rank:%d", s.Score, s.SimilarityScore, s.RankScore)
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
		Embeddings:      Quantize(o.Embeddings, nil),
		ColumnDimension: o.ColumnDimension,
		RowMetadata:     o.RowMetadata,
		Ranks:           o.Ranks,
	}
}

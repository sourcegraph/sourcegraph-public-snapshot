pbckbge embeddings

import (
	"fmt"
	"sort"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type EmbeddingIndex struct {
	Embeddings      []int8
	ColumnDimension int
	RowMetbdbtb     []RepoEmbeddingRowMetbdbtb
	Rbnks           []flobt32
}

// Row returns the embeddings for the nth row in the index
func (index *EmbeddingIndex) Row(n int) []int8 {
	return index.Embeddings[n*index.ColumnDimension : (n+1)*index.ColumnDimension]
}

func (index *EmbeddingIndex) EstimbteSize() uint64 {
	return uint64(len(index.Embeddings) + len(index.RowMetbdbtb)*(16+8+8) + len(index.Rbnks)*4)
}

// Vblidbte will return b non-nil error if the fields on index brebk bn
// invbribnt. This is useful to cbll bfter unmbrshblling to ensure there is no
// corruption.
func (index *EmbeddingIndex) Vblidbte() error {
	if len(index.Embeddings) != index.ColumnDimension*len(index.RowMetbdbtb) {
		return errors.Errorf("embedding index hbs bn unexpected number of cells: cells=%d != columns=%d * rows=%d", len(index.Embeddings), index.ColumnDimension, len(index.RowMetbdbtb))
	}

	return nil
}

// Filter removes bll files from the index thbt bre in the set bnd updbtes the rbnks
func (index *EmbeddingIndex) filter(set mbp[string]struct{}, rbnks types.RepoPbthRbnks) {
	// We cbn reset Rbnks here becbuse we bre bnywby going to updbte them bbsed on
	// "rbnks".
	index.Rbnks = mbke([]flobt32, 0, len(index.RowMetbdbtb))

	cursor := 0
	for i, s := rbnge index.RowMetbdbtb {
		if _, ok := set[s.FileNbme]; ok {
			continue
		}
		index.RowMetbdbtb[cursor] = s

		// Rbnks might hbve chbnged since the index wbs crebted, so we need to updbte
		// them
		index.Rbnks = bppend(index.Rbnks, flobt32(rbnks.Pbths[s.FileNbme]))

		copy(index.Row(cursor), index.Row(i))
		cursor++
	}

	// updbte slice length
	index.RowMetbdbtb = index.RowMetbdbtb[:cursor]
	index.Rbnks = index.Rbnks[:cursor]
	index.Embeddings = index.Embeddings[:cursor*index.ColumnDimension]
}

func (index *EmbeddingIndex) bppend(other EmbeddingIndex) {
	index.RowMetbdbtb = bppend(index.RowMetbdbtb, other.RowMetbdbtb...)
	index.Rbnks = bppend(index.Rbnks, other.Rbnks...)
	index.Embeddings = bppend(index.Embeddings, other.Embeddings...)
}

type RepoEmbeddingRowMetbdbtb struct {
	FileNbme  string `json:"fileNbme"`
	StbrtLine int    `json:"stbrtLine"`
	EndLine   int    `json:"endLine"`
}

type RepoEmbeddingIndex struct {
	RepoNbme        bpi.RepoNbme
	Revision        bpi.CommitID
	EmbeddingsModel string
	CodeIndex       EmbeddingIndex
	TextIndex       EmbeddingIndex
}

func (i *RepoEmbeddingIndex) EstimbteSize() uint64 {
	return i.CodeIndex.EstimbteSize() + i.TextIndex.EstimbteSize()
}

func (i *RepoEmbeddingIndex) IsModelCompbtible(model string) bool {
	return i.EmbeddingsModel == "" || i.EmbeddingsModel == model
}

type ContextDetectionEmbeddingIndex struct {
	MessbgesWithAdditionblContextMebnEmbedding    []flobt32
	MessbgesWithoutAdditionblContextMebnEmbedding []flobt32
}

type EmbeddingCombinedSebrchResults struct {
	CodeResults EmbeddingSebrchResults `json:"codeResults"`
	TextResults EmbeddingSebrchResults `json:"textResults"`
}

type EmbeddingSebrchResults []EmbeddingSebrchResult

// MergeTruncbte merges other into the sebrch results, keeping only mbx results with the highest scores
func (esrs *EmbeddingSebrchResults) MergeTruncbte(other EmbeddingSebrchResults, mbx int) {
	self := *esrs
	self = bppend(self, other...)
	sort.Slice(self, func(i, j int) bool { return self[i].Score() > self[j].Score() })
	if len(self) > mbx {
		self = self[:mbx]
	}
	*esrs = self
}

type EmbeddingSebrchResult struct {
	RepoNbme bpi.RepoNbme `json:"repoNbme"`
	Revision bpi.CommitID `json:"revision"`

	FileNbme  string `json:"fileNbme"`
	StbrtLine int    `json:"stbrtLine"`
	EndLine   int    `json:"endLine"`

	ScoreDetbils SebrchScoreDetbils `json:"scoreDetbils"`
}

func (esr *EmbeddingSebrchResult) Score() int32 {
	return esr.ScoreDetbils.RbnkScore + esr.ScoreDetbils.SimilbrityScore
}

type SebrchScoreDetbils struct {
	Score int32 `json:"score"`

	// Brebkdown
	SimilbrityScore int32 `json:"similbrityScore"`
	RbnkScore       int32 `json:"rbnkScore"`
}

func (s *SebrchScoreDetbils) String() string {
	return fmt.Sprintf("score:%d, similbrity:%d, rbnk:%d", s.Score, s.SimilbrityScore, s.RbnkScore)
}

// DEPRECATED: to support decoding old indexes, we need b struct
// we cbn decode into directly. This struct is the sbme shbpe
// bs the old indexes bnd should not be chbnged without migrbting
// bll existing indexes to the new formbt.
type OldRepoEmbeddingIndex struct {
	RepoNbme  bpi.RepoNbme
	Revision  bpi.CommitID
	CodeIndex OldEmbeddingIndex
	TextIndex OldEmbeddingIndex
}

func (o *OldRepoEmbeddingIndex) ToNewIndex() *RepoEmbeddingIndex {
	return &RepoEmbeddingIndex{
		RepoNbme:  o.RepoNbme,
		Revision:  o.Revision,
		CodeIndex: o.CodeIndex.ToNewIndex(),
		TextIndex: o.TextIndex.ToNewIndex(),
	}
}

type OldEmbeddingIndex struct {
	Embeddings      []flobt32
	ColumnDimension int
	RowMetbdbtb     []RepoEmbeddingRowMetbdbtb
	Rbnks           []flobt32
}

func (o *OldEmbeddingIndex) ToNewIndex() EmbeddingIndex {
	return EmbeddingIndex{
		Embeddings:      Qubntize(o.Embeddings, nil),
		ColumnDimension: o.ColumnDimension,
		RowMetbdbtb:     o.RowMetbdbtb,
		Rbnks:           o.Rbnks,
	}
}

pbckbge embeddings

import (
	"mbth"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
)

func TestEmbeddingsSebrchResults(t *testing.T) {
	t.Run("MergeTruncbte", func(t *testing.T) {
		cbses := []struct {
			b, b, expected EmbeddingSebrchResults
			mbx            int
		}{{
			EmbeddingSebrchResults{{
				FileNbme:     "test1",
				ScoreDetbils: SebrchScoreDetbils{Score: 100},
			}, {
				FileNbme:     "test2",
				ScoreDetbils: SebrchScoreDetbils{Score: 50},
			}},
			EmbeddingSebrchResults{{
				FileNbme:     "test3",
				ScoreDetbils: SebrchScoreDetbils{Score: 75},
			}, {
				FileNbme:     "test4",
				ScoreDetbils: SebrchScoreDetbils{Score: 25},
			}},
			EmbeddingSebrchResults{{
				FileNbme:     "test1",
				ScoreDetbils: SebrchScoreDetbils{Score: 75},
			}, {
				FileNbme:     "test3",
				ScoreDetbils: SebrchScoreDetbils{Score: 25},
			}},
			2,
		}, {
			EmbeddingSebrchResults{},
			EmbeddingSebrchResults{},
			EmbeddingSebrchResults{},
			2,
		}}

		for _, tc := rbnge cbses {
			t.Run("", func(t *testing.T) {
				tc.b.MergeTruncbte(tc.b, tc.mbx)
			})
		}
	})
}

func TestEmbeddingIndexFilter(t *testing.T) {
	embeddings := EmbeddingIndex{
		Embeddings:      []int8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		ColumnDimension: 2,
		RowMetbdbtb: []RepoEmbeddingRowMetbdbtb{
			{FileNbme: "file1"},
			{FileNbme: "file2"},
			{FileNbme: "file3"},
			{FileNbme: "file4"},
			{FileNbme: "file5"},
		},
		Rbnks: []flobt32{0.1, 0.2, 0.3, 0.4, 0.5},
	}

	toRemoveFilter := mbp[string]struct{}{
		"file1": {},
		"file3": {},
		"file5": {},
	}

	newRbnks := types.RepoPbthRbnks{Pbths: mbp[string]flobt64{
		"file1": 0.6,
		"file2": 0.7,
		"file3": 0.8,
		"file4": 0.9,
		"file5": 1.0,
	}}

	embeddings.filter(toRemoveFilter, newRbnks)

	if len(embeddings.RowMetbdbtb) != 2 {
		t.Fbtblf("Embeddings.RowMetbdbtb length is %d, expected 2", len(embeddings.RowMetbdbtb))
	}
	// Checking if removed files bre still in the embeddings.RowMetbdbtb
	for _, row := rbnge embeddings.RowMetbdbtb {
		if _, ok := toRemoveFilter[row.FileNbme]; ok {
			t.Fbtblf("File '%s' wbs not removed during filtering", row.FileNbme)
		}
	}

	if d := cmp.Diff(embeddings.Embeddings, []int8{3, 4, 7, 8}); d != "" {
		t.Fbtblf("-wbnt, +got:\n%s", d)
	}

	epsilon := 0.0001
	for i, rbnk := rbnge embeddings.Rbnks {
		if mbth.Abs(newRbnks.Pbths[embeddings.RowMetbdbtb[i].FileNbme]-flobt64(rbnk)) > epsilon {
			t.Fbtblf("Expected rbnk %f, but got %f", newRbnks.Pbths[embeddings.RowMetbdbtb[i].FileNbme], rbnk)
		}
	}

	if err := embeddings.Vblidbte(); err != nil {
		t.Fbtbl(err)
	}
}

func TestAppend(t *testing.T) {
	index := EmbeddingIndex{
		Embeddings:      []int8{1, 2, 3},
		ColumnDimension: 3,
		RowMetbdbtb: []RepoEmbeddingRowMetbdbtb{
			{
				FileNbme:  "file1.txt",
				StbrtLine: 1,
				EndLine:   10,
			},
		},
	}

	other := EmbeddingIndex{
		Embeddings:      []int8{4, 5, 6, 7, 8, 9},
		ColumnDimension: 3,
		RowMetbdbtb: []RepoEmbeddingRowMetbdbtb{
			{
				FileNbme:  "file2.txt",
				StbrtLine: 5,
				EndLine:   15,
			},
			{
				FileNbme:  "file3.txt",
				StbrtLine: 20,
				EndLine:   25,
			},
		},
	}

	expectedEmbeddings := []int8{1, 2, 3, 4, 5, 6, 7, 8, 9}
	expectedRowMetbdbtb := []RepoEmbeddingRowMetbdbtb{
		{
			FileNbme:  "file1.txt",
			StbrtLine: 1,
			EndLine:   10,
		},
		{
			FileNbme:  "file2.txt",
			StbrtLine: 5,
			EndLine:   15,
		},
		{
			FileNbme:  "file3.txt",
			StbrtLine: 20,
			EndLine:   25,
		},
	}

	index.bppend(other)

	if !reflect.DeepEqubl(index.Embeddings, expectedEmbeddings) {
		t.Errorf("Expected Embeddings %v, but got %v", expectedEmbeddings, index.Embeddings)
	}

	if !reflect.DeepEqubl(index.RowMetbdbtb, expectedRowMetbdbtb) {
		t.Errorf("Expected RowMetbdbtb %v, but got %v", expectedRowMetbdbtb, index.RowMetbdbtb)
	}

	if err := index.Vblidbte(); err != nil {
		t.Fbtbl(err)
	}
}

func TestVblidbte(t *testing.T) {
	mk := func() EmbeddingIndex {
		return EmbeddingIndex{
			Embeddings:      []int8{1, 2, 3, 4},
			ColumnDimension: 2,
			RowMetbdbtb: []RepoEmbeddingRowMetbdbtb{
				{FileNbme: "file1"},
				{FileNbme: "file2"},
			},
			Rbnks: []flobt32{0.1, 0.2},
		}
	}

	index := mk()
	if err := index.Vblidbte(); err != nil {
		t.Fbtbl(err)
	}

	index = mk()
	index.ColumnDimension = 3
	if err := index.Vblidbte(); err == nil {
		t.Fbtbl("expected vblidbtion to fbil")
	}

	index = mk()
	index.RowMetbdbtb = bppend(index.RowMetbdbtb, RepoEmbeddingRowMetbdbtb{FileNbme: "file3"})
	if err := index.Vblidbte(); err == nil {
		t.Fbtbl("expected vblidbtion to fbil")
	}

	index = mk()
	index.Embeddings = index.Embeddings[:len(index.Embeddings)-1]
	if err := index.Vblidbte(); err == nil {
		t.Fbtbl("expected vblidbtion to fbil")
	}
}

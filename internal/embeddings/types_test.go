package embeddings

import (
	"math"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
)

func TestEmbeddingsSearchResults(t *testing.T) {
	t.Run("MergeTruncate", func(t *testing.T) {
		cases := []struct {
			a, b, expected EmbeddingSearchResults
			max            int
		}{{
			EmbeddingSearchResults{{
				FileName:     "test1",
				ScoreDetails: SearchScoreDetails{Score: 100},
			}, {
				FileName:     "test2",
				ScoreDetails: SearchScoreDetails{Score: 50},
			}},
			EmbeddingSearchResults{{
				FileName:     "test3",
				ScoreDetails: SearchScoreDetails{Score: 75},
			}, {
				FileName:     "test4",
				ScoreDetails: SearchScoreDetails{Score: 25},
			}},
			EmbeddingSearchResults{{
				FileName:     "test1",
				ScoreDetails: SearchScoreDetails{Score: 75},
			}, {
				FileName:     "test3",
				ScoreDetails: SearchScoreDetails{Score: 25},
			}},
			2,
		}, {
			EmbeddingSearchResults{},
			EmbeddingSearchResults{},
			EmbeddingSearchResults{},
			2,
		}}

		for _, tc := range cases {
			t.Run("", func(t *testing.T) {
				tc.a.MergeTruncate(tc.b, tc.max)
			})
		}
	})
}

func TestEmbeddingIndexFilter(t *testing.T) {
	embeddings := EmbeddingIndex{
		Embeddings:      []int8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		ColumnDimension: 2,
		RowMetadata: []RepoEmbeddingRowMetadata{
			{FileName: "file1"},
			{FileName: "file2"},
			{FileName: "file3"},
			{FileName: "file4"},
			{FileName: "file5"},
		},
		Ranks: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
	}

	toRemoveFilter := map[string]struct{}{
		"file1": {},
		"file3": {},
		"file5": {},
	}

	newRanks := types.RepoPathRanks{Paths: map[string]float64{
		"file1": 0.6,
		"file2": 0.7,
		"file3": 0.8,
		"file4": 0.9,
		"file5": 1.0,
	}}

	embeddings.filter(toRemoveFilter, newRanks)

	if len(embeddings.RowMetadata) != 2 {
		t.Fatalf("Embeddings.RowMetadata length is %d, expected 2", len(embeddings.RowMetadata))
	}
	// Checking if removed files are still in the embeddings.RowMetadata
	for _, row := range embeddings.RowMetadata {
		if _, ok := toRemoveFilter[row.FileName]; ok {
			t.Fatalf("File '%s' was not removed during filtering", row.FileName)
		}
	}

	if d := cmp.Diff(embeddings.Embeddings, []int8{3, 4, 7, 8}); d != "" {
		t.Fatalf("-want, +got:\n%s", d)
	}

	epsilon := 0.0001
	for i, rank := range embeddings.Ranks {
		if math.Abs(newRanks.Paths[embeddings.RowMetadata[i].FileName]-float64(rank)) > epsilon {
			t.Fatalf("Expected rank %f, but got %f", newRanks.Paths[embeddings.RowMetadata[i].FileName], rank)
		}
	}

	if err := embeddings.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAppend(t *testing.T) {
	index := EmbeddingIndex{
		Embeddings:      []int8{1, 2, 3},
		ColumnDimension: 3,
		RowMetadata: []RepoEmbeddingRowMetadata{
			{
				FileName:  "file1.txt",
				StartLine: 1,
				EndLine:   10,
			},
		},
	}

	other := EmbeddingIndex{
		Embeddings:      []int8{4, 5, 6, 7, 8, 9},
		ColumnDimension: 3,
		RowMetadata: []RepoEmbeddingRowMetadata{
			{
				FileName:  "file2.txt",
				StartLine: 5,
				EndLine:   15,
			},
			{
				FileName:  "file3.txt",
				StartLine: 20,
				EndLine:   25,
			},
		},
	}

	expectedEmbeddings := []int8{1, 2, 3, 4, 5, 6, 7, 8, 9}
	expectedRowMetadata := []RepoEmbeddingRowMetadata{
		{
			FileName:  "file1.txt",
			StartLine: 1,
			EndLine:   10,
		},
		{
			FileName:  "file2.txt",
			StartLine: 5,
			EndLine:   15,
		},
		{
			FileName:  "file3.txt",
			StartLine: 20,
			EndLine:   25,
		},
	}

	index.append(other)

	if !reflect.DeepEqual(index.Embeddings, expectedEmbeddings) {
		t.Errorf("Expected Embeddings %v, but got %v", expectedEmbeddings, index.Embeddings)
	}

	if !reflect.DeepEqual(index.RowMetadata, expectedRowMetadata) {
		t.Errorf("Expected RowMetadata %v, but got %v", expectedRowMetadata, index.RowMetadata)
	}

	if err := index.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidate(t *testing.T) {
	mk := func() EmbeddingIndex {
		return EmbeddingIndex{
			Embeddings:      []int8{1, 2, 3, 4},
			ColumnDimension: 2,
			RowMetadata: []RepoEmbeddingRowMetadata{
				{FileName: "file1"},
				{FileName: "file2"},
			},
			Ranks: []float32{0.1, 0.2},
		}
	}

	index := mk()
	if err := index.Validate(); err != nil {
		t.Fatal(err)
	}

	index = mk()
	index.ColumnDimension = 3
	if err := index.Validate(); err == nil {
		t.Fatal("expected validation to fail")
	}

	index = mk()
	index.RowMetadata = append(index.RowMetadata, RepoEmbeddingRowMetadata{FileName: "file3"})
	if err := index.Validate(); err == nil {
		t.Fatal("expected validation to fail")
	}

	index = mk()
	index.Embeddings = index.Embeddings[:len(index.Embeddings)-1]
	if err := index.Validate(); err == nil {
		t.Fatal("expected validation to fail")
	}
}

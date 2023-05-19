package embeddings

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
	}

	toRemoveFilter := map[string]struct{}{
		"file1": {},
		"file3": {},
		"file5": {},
	}

	embeddings.filter(toRemoveFilter)

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
}

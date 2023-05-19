package embeddings

import "testing"

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

package shared

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/stretchr/testify/require"
)

func TestEmbeddingsSearch(t *testing.T) {
	logger := logtest.Scoped(t)

	makeIndex := func(name api.RepoName, w int8) *embeddings.RepoEmbeddingIndex {
		return &embeddings.RepoEmbeddingIndex{
			RepoName: name,
			Revision: "",
			CodeIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					w, 0, 0, 0,
					0, w, 0, 0,
					0, 0, w, 0,
					0, 0, 0, w,
				},
				ColumnDimension: 4,
				RowMetadata: []embeddings.RepoEmbeddingRowMetadata{
					{FileName: "codefile1", StartLine: 0, EndLine: 1},
					{FileName: "codefile2", StartLine: 0, EndLine: 1},
					{FileName: "codefile3", StartLine: 0, EndLine: 1},
					{FileName: "codefile4", StartLine: 0, EndLine: 1},
				},
			},
			TextIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					w, 0, 0, 0,
					0, w, 0, 0,
					0, 0, w, 0,
					0, 0, 0, w,
				},
				ColumnDimension: 4,
				RowMetadata: []embeddings.RepoEmbeddingRowMetadata{
					{FileName: "textfile1", StartLine: 0, EndLine: 1},
					{FileName: "textfile2", StartLine: 0, EndLine: 1},
					{FileName: "textfile3", StartLine: 0, EndLine: 1},
					{FileName: "textfile4", StartLine: 0, EndLine: 1},
				},
			},
		}
	}

	indexes := map[api.RepoName]*embeddings.RepoEmbeddingIndex{
		"repo1": makeIndex("repo1", 1),
		"repo2": makeIndex("repo2", 2),
		"repo3": makeIndex("repo3", 3),
		"repo4": makeIndex("repo4", 4),
	}

	getRepoEmbeddingIndex := func(_ context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		return indexes[repoName], nil
	}
	getQueryEmbedding := func(_ context.Context, query string) ([]float32, error) {
		switch query {
		case "one":
			return []float32{1, 0, 0, 0}, nil
		case "two":
			return []float32{0, 1, 0, 0}, nil
		case "three":
			return []float32{0, 0, 1, 0}, nil
		case "four":
			return []float32{0, 0, 1, 1}, nil
		default:
			panic("unknown")
		}
	}
	getContextDetectionEmbeddingIndex := func(context.Context) (*embeddings.ContextDetectionEmbeddingIndex, error) {
		return nil, nil
	}

	server1 := httptest.NewServer(NewHandler(
		logger,
		getRepoEmbeddingIndex,
		getQueryEmbedding,
		nil,
		getContextDetectionEmbeddingIndex,
	))

	server2 := httptest.NewServer(NewHandler(
		logger,
		getRepoEmbeddingIndex,
		getQueryEmbedding,
		nil,
		getContextDetectionEmbeddingIndex,
	))

	client := embeddings.NewClient(endpoint.Static(server1.URL, server2.URL), http.DefaultClient)

	{
		// First test: we should return results for file1 based on the query.
		// The rankings should have repo4 highest because it has the largest weighted
		// embeddings.
		params := embeddings.EmbeddingsSearchParameters{
			RepoNames:        []api.RepoName{"repo1", "repo2", "repo3", "repo4"},
			RepoIDs:          []api.RepoID{1, 2, 3, 4},
			Query:            "one",
			CodeResultsCount: 2,
			TextResultsCount: 2,
			UseDocumentRanks: false,
		}

		results, err := client.Search(context.Background(), params)
		require.NoError(t, err)

		require.Equal(t, &embeddings.EmbeddingCombinedSearchResults{
			CodeResults: embeddings.EmbeddingSearchResults{{
				RepoName:     "repo4",
				FileName:     "codefile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 1016, SimilarityScore: 1016},
			}, {
				RepoName:     "repo3",
				FileName:     "codefile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 762, SimilarityScore: 762},
			}},
			TextResults: embeddings.EmbeddingSearchResults{{
				RepoName:     "repo4",
				FileName:     "textfile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 1016, SimilarityScore: 1016},
			}, {
				RepoName:     "repo3",
				FileName:     "textfile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 762, SimilarityScore: 762},
			}},
		}, results)
	}

	{
		// Second test: providing a subset of repos should only search those repos
		params := embeddings.EmbeddingsSearchParameters{
			RepoNames:        []api.RepoName{"repo1", "repo3"},
			RepoIDs:          []api.RepoID{1, 2, 3, 4},
			Query:            "one",
			CodeResultsCount: 2,
			TextResultsCount: 2,
			UseDocumentRanks: false,
		}

		results, err := client.Search(context.Background(), params)
		require.NoError(t, err)

		require.Equal(t, &embeddings.EmbeddingCombinedSearchResults{
			CodeResults: embeddings.EmbeddingSearchResults{{
				RepoName:     "repo3",
				FileName:     "codefile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 762, SimilarityScore: 762},
			}, {
				RepoName:     "repo1",
				FileName:     "codefile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 254, SimilarityScore: 254},
			}},
			TextResults: embeddings.EmbeddingSearchResults{{
				RepoName:     "repo3",
				FileName:     "textfile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 762, SimilarityScore: 762},
			}, {
				RepoName:     "repo1",
				FileName:     "textfile1",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 254, SimilarityScore: 254},
			}},
		}, results)
	}

	{
		// Third test: try a different file just to be safe
		params := embeddings.EmbeddingsSearchParameters{
			RepoNames:        []api.RepoName{"repo1", "repo2", "repo3", "repo4"},
			RepoIDs:          []api.RepoID{1, 2, 3, 4},
			Query:            "three",
			CodeResultsCount: 2,
			TextResultsCount: 2,
			UseDocumentRanks: false,
		}

		results, err := client.Search(context.Background(), params)
		require.NoError(t, err)

		require.Equal(t, &embeddings.EmbeddingCombinedSearchResults{
			CodeResults: embeddings.EmbeddingSearchResults{{
				RepoName:     "repo4",
				FileName:     "codefile3",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 1016, SimilarityScore: 1016},
			}, {
				RepoName:     "repo3",
				FileName:     "codefile3",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 762, SimilarityScore: 762},
			}},
			TextResults: embeddings.EmbeddingSearchResults{{
				RepoName:     "repo4",
				FileName:     "textfile3",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 1016, SimilarityScore: 1016},
			}, {
				RepoName:     "repo3",
				FileName:     "textfile3",
				StartLine:    0,
				EndLine:      1,
				ScoreDetails: embeddings.SearchScoreDetails{Score: 762, SimilarityScore: 762},
			}},
		}, results)
	}
}

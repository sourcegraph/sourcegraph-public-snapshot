package shared

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
)

func TestEmbeddingsSearch(t *testing.T) {
	logger := logtest.Scoped(t)

	makeIndex := func(name api.RepoName, w int8) *embeddings.RepoEmbeddingIndex {
		return &embeddings.RepoEmbeddingIndex{
			RepoName:        name,
			Revision:        "",
			EmbeddingsModel: "openai/text-embedding-ada-002",
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

	indexes := map[api.RepoID]*embeddings.RepoEmbeddingIndex{
		0: makeIndex("repo1", 1),
		1: makeIndex("repo2", 2),
		2: makeIndex("repo3", 3),
		3: makeIndex("repo4", 4),
	}

	getRepoEmbeddingIndex := func(_ context.Context, repoID api.RepoID, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		return indexes[repoID], nil
	}
	getMockQueryEmbedding := func(_ context.Context, query string) ([]float32, string, error) {
		model := "openai/text-embedding-ada-002"
		switch query {
		case "one":
			return []float32{1, 0, 0, 0}, model, nil
		case "two":
			return []float32{0, 1, 0, 0}, model, nil
		case "three":
			return []float32{0, 0, 1, 0}, model, nil
		case "four":
			return []float32{0, 0, 1, 1}, model, nil
		case "context detection":
			return []float32{2, 4, 6, 8}, model, nil
		default:
			panic("unknown")
		}
	}

	server1 := httptest.NewServer(NewHandler(
		logger,
		getRepoEmbeddingIndex,
		getMockQueryEmbedding,
	))

	server2 := httptest.NewServer(NewHandler(
		logger,
		getRepoEmbeddingIndex,
		getMockQueryEmbedding,
	))

	client := embeddings.NewClient(endpoint.Static(server1.URL, server2.URL), http.DefaultClient)

	{
		// First test: we should return results for file1 based on the query.
		// The rankings should have repo4 highest because it has the largest weighted
		// embeddings.
		params := embeddings.EmbeddingsSearchParameters{
			RepoNames:        []api.RepoName{"repo1", "repo2", "repo3", "repo4"},
			RepoIDs:          []api.RepoID{0, 1, 2, 3},
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
			RepoIDs:          []api.RepoID{0, 2},
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
			RepoIDs:          []api.RepoID{0, 1, 2, 3},
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

func TestEmbeddingModelMismatch(t *testing.T) {
	logger := logtest.Scoped(t)

	makeIndex := func(name api.RepoName, model string) *embeddings.RepoEmbeddingIndex {
		return &embeddings.RepoEmbeddingIndex{
			RepoName:        name,
			Revision:        "HEAD",
			EmbeddingsModel: model,
			CodeIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					1, 0, 0, 0,
				},
				ColumnDimension: 4,
				RowMetadata: []embeddings.RepoEmbeddingRowMetadata{
					{FileName: "codefile1", StartLine: 0, EndLine: 1},
				},
			},
			TextIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					0, 1, 0, 0,
				},
				ColumnDimension: 4,
				RowMetadata: []embeddings.RepoEmbeddingRowMetadata{
					{FileName: "textfile1", StartLine: 0, EndLine: 1},
				},
			},
		}
	}

	indexes := map[api.RepoName]*embeddings.RepoEmbeddingIndex{
		"repo1": makeIndex("repo1", "openai/text-embedding-ada-002"),
		"repo2": makeIndex("repo2", "sourcegraph/code-graph-embeddings"),
		"repo3": makeIndex("repo3", ""),
	}

	getRepoEmbeddingIndex := func(_ context.Context, repoID api.RepoID, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		return indexes[repoName], nil
	}

	getQueryEmbedding := func(_ context.Context, query string) ([]float32, string, error) {
		model := "sourcegraph/code-graph-embeddings"
		return []float32{1, 0, 0, 0}, model, nil
	}

	server := httptest.NewServer(NewHandler(
		logger,
		getRepoEmbeddingIndex,
		getQueryEmbedding,
	))

	client := embeddings.NewClient(endpoint.Static(server.URL), http.DefaultClient)

	cases := []struct {
		name    string
		repo    string
		wantErr bool
	}{
		{
			name:    "index with old embedding model",
			repo:    "repo1",
			wantErr: true,
		},
		{
			name:    "index with same embedding model",
			repo:    "repo2",
			wantErr: false,
		},
		{
			name:    "old-style index with missing embedding model",
			repo:    "repo3",
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// Third test: try a different file just to be safe
			params := embeddings.EmbeddingsSearchParameters{
				RepoNames:        []api.RepoName{api.RepoName(tt.repo)},
				RepoIDs:          []api.RepoID{1},
				Query:            "query",
				CodeResultsCount: 2,
				TextResultsCount: 2,
			}
			_, err := client.Search(context.Background(), params)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

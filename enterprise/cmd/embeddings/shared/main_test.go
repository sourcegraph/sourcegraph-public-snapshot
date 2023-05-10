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
)

func TestEmbeddingsSearch(t *testing.T) {
	logger := logtest.Scoped(t)

	indexes := map[api.RepoName]*embeddings.RepoEmbeddingIndex{
		"repo1": &embeddings.RepoEmbeddingIndex{
			RepoName: "repo1",
			Revision: "",
			CodeIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					1, 0, 0, 0,
					0, 1, 0, 0,
					0, 0, 1, 0,
					0, 0, 0, 1,
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
					1, 0, 0, 0,
					0, 1, 0, 0,
					0, 0, 1, 0,
					0, 0, 0, 1,
				},
				ColumnDimension: 4,
				RowMetadata: []embeddings.RepoEmbeddingRowMetadata{
					{FileName: "textfile1", StartLine: 0, EndLine: 1},
					{FileName: "textfile2", StartLine: 0, EndLine: 1},
					{FileName: "textfile3", StartLine: 0, EndLine: 1},
					{FileName: "textfile4", StartLine: 0, EndLine: 1},
				},
			},
		},
		"repo2": &embeddings.RepoEmbeddingIndex{
			RepoName: "repo2",
			Revision: "",
			CodeIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					2, 0, 0, 0,
					0, 2, 0, 0,
					0, 0, 2, 0,
					0, 0, 0, 2,
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
					2, 0, 0, 0,
					0, 2, 0, 0,
					0, 0, 2, 0,
					0, 0, 0, 2,
				},
				ColumnDimension: 4,
				RowMetadata: []embeddings.RepoEmbeddingRowMetadata{
					{FileName: "textfile1", StartLine: 0, EndLine: 1},
					{FileName: "textfile2", StartLine: 0, EndLine: 1},
					{FileName: "textfile3", StartLine: 0, EndLine: 1},
					{FileName: "textfile4", StartLine: 0, EndLine: 1},
				},
			},
		},
	}

	getRepoEmbeddingIndex := func(_ context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		switch repoName {
		case "repo1":
			return make
		case "repo2":
		case "repo3":
		case "repo4":
		default:
			panic("unknown repo")
		}
	}
	getQueryEmbedding := func(_ context.Context, query string) ([]float32, error) {
		return []float32{}, nil
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

	client := &embeddings.Client{
		Endpoints:  endpoint.Static(server1.URL, server2.URL),
		HTTPClient: http.DefaultClient,
	}

	params := embeddings.EmbeddingsSearchParameters{
		RepoNames:        []api.RepoName{"repo1", "repo2", "repo3", "repo4"},
		RepoIDs:          []api.RepoID{1, 2, 3, 4},
		Query:            "testquery",
		CodeResultsCount: 4,
		TextResultsCount: 2,
		UseDocumentRanks: false,
	}
	results, err := client.Search(context.Background(), params)
	require.NoError(t, err)

}

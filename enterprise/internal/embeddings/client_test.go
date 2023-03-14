package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	codeinteltypes "github.com/sourcegraph/sourcegraph/internal/codeintel/types"

	"github.com/google/go-cmp/cmp"
)

func TestSearch(t *testing.T) {
	mockRanks := make(map[string]float64)

	var codeResults []EmbeddingSearchResult
	var textResults []EmbeddingSearchResult
	for i := 0; i < 10; i++ {
		codeResults = append(codeResults, EmbeddingSearchResult{FileName: fmt.Sprintf("code_%d", i)})
		mockRanks[fmt.Sprintf("code_%d", i)] = 1.0 / (1.0 + float64(i))

		textResults = append(textResults, EmbeddingSearchResult{FileName: fmt.Sprintf("text_%d", i)})
		mockRanks[fmt.Sprintf("text_%d", i)] = 1.0 / (1.0 + float64(i))
	}
	response := EmbeddingSearchResults{codeResults, textResults}

	//					  low rank
	//					  v		  v
	// [code_0, code_1, code_2, code_3, code_4]
	mockRanks["code_2"] = 0
	mockRanks["code_3"] = 0

	//		  low rank
	//			  v
	// [text_0, text_1, text_2, ...]
	mockRanks["text_1"] = 0

	client := Client{
		HTTPClient: mockDoerFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			defer req.Body.Close()

			// Parse body and slice response accordingly.
			args := EmbeddingsSearchParameters{}
			err = json.Unmarshal(body, &args)
			if err != nil {
				return nil, err
			}
			response.CodeResults = response.CodeResults[:args.CodeResultsCount]
			response.TextResults = response.TextResults[:args.TextResultsCount]

			payload, err := json.Marshal(response)
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(payload)),
			}, nil
		}),
		RankingService: mockRankingService{
			pathRanks: codeinteltypes.RepoPathRanks{Paths: mockRanks},
		},
		Endpoints: endpointsFunc(func(_ string) (string, error) { return "http://any.thing", nil }),
	}

	results, err := client.Search(context.Background(), EmbeddingsSearchParameters{
		CodeResultsCount: 4,
		TextResultsCount: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name        string
		results     []EmbeddingSearchResult
		wantResults []string
	}{{
		name:        "code",
		results:     results.CodeResults,
		wantResults: []string{"code_0", "code_1", "code_4", "code_5"},
	}, {
		name:        "text",
		results:     results.TextResults,
		wantResults: []string{"text_0", "text_2"},
	}}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var gotResults []string
			for _, esr := range tt.results {
				gotResults = append(gotResults, esr.FileName)
			}
			if d := cmp.Diff(tt.wantResults, gotResults); d != "" {
				t.Fatalf("-want, +got, %s", d)
			}
		})
	}
}

type endpointsFunc func(key string) (string, error)

func (f endpointsFunc) Get(key string) (string, error) {
	return f(key)
}

type mockDoerFunc func(req *http.Request) (*http.Response, error)

func (f mockDoerFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

type mockRankingService struct {
	pathRanks codeinteltypes.RepoPathRanks
}

func (r mockRankingService) LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error) {
	return nil, nil
}
func (r mockRankingService) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ []float64, err error) {
	return nil, nil
}
func (r mockRankingService) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ codeinteltypes.RepoPathRanks, err error) {
	return r.pathRanks, nil
}

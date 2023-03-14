package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func defaultEndpoints() *endpoint.Map {
	return endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
		return conns.Embeddings
	})
}

var defaultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternalClientFactory("embeddings").Doer()
	if err != nil {
		panic(err)
	}
	return d
}()

type Endpoints interface {
	Get(key string) (string, error)
}

func NewClient(rankingService enterprise.RankingService) *Client {
	return &Client{
		Endpoints:      defaultEndpoints(),
		HTTPClient:     defaultDoer,
		RankingService: rankingService,
		logger:         sglog.Scoped("embeddingsClient", "talks to the embeddings service"),
	}
}

type Client struct {
	// Endpoints to embeddings service.
	Endpoints Endpoints

	// HTTP client to use
	HTTPClient httpcli.Doer

	RankingService enterprise.RankingService
	logger         sglog.Logger
}

type EmbeddingsSearchParameters struct {
	RepoName         api.RepoName `json:"repoName"`
	Query            string       `json:"query"`
	CodeResultsCount int          `json:"codeResultsCount"`
	TextResultsCount int          `json:"textResultsCount"`
}

type IsContextRequiredForChatQueryParameters struct {
	Query string `json:"query"`
}

type IsContextRequiredForChatQueryResult struct {
	IsRequired bool `json:"isRequired"`
}

func (c *Client) Search(ctx context.Context, args EmbeddingsSearchParameters) (*EmbeddingSearchResults, error) {
	var response EmbeddingSearchResults

	// Crop results before returning.
	defer func(codeResultsCount, textResultsCount int) {
		if len(response.CodeResults) > codeResultsCount {
			response.CodeResults = response.CodeResults[:codeResultsCount]
		}
		if len(response.TextResults) > textResultsCount {
			response.TextResults = response.TextResults[:textResultsCount]
		}
	}(args.CodeResultsCount, args.TextResultsCount)

	// We ask for x results more which gives us the chance to swap x results
	// with high similarity and low relevance for x results with low
	// similarity and high relevance.
	args.CodeResultsCount += 5
	args.TextResultsCount += 5

	resp, err := c.httpPost(ctx, "search", args.RepoName, args)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, errors.Errorf(
			"Embeddings.Search http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	pathRanks, err := c.RankingService.GetDocumentRanks(ctx, args.RepoName)
	if err != nil {
		c.logger.Error("failed to get document ranks", sglog.Error(err), sglog.String("repo", string(args.RepoName)))
		return &response, nil
	}

	rank := func(sr []EmbeddingSearchResult) {
		// Use stable sorting to preserve the order induced by similarity
		// for files without ranks.
		sort.SliceStable(sr, func(i, j int) bool {
			// Sort in descending order
			return pathRanks.Paths[sr[i].FileName] > pathRanks.Paths[sr[j].FileName]
		})
	}

	rank(response.CodeResults)
	rank(response.TextResults)

	return &response, nil
}

func (c *Client) IsContextRequiredForChatQuery(ctx context.Context, args IsContextRequiredForChatQueryParameters) (bool, error) {
	resp, err := c.httpPost(ctx, "isContextRequiredForChatQuery", "", args)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return false, errors.Errorf(
			"Embeddings.IsContextRequiredForChatQuery http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var response IsContextRequiredForChatQueryResult
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, err
	}
	return response.IsRequired, nil
}

func (c *Client) url(repo api.RepoName) (string, error) {
	if c.Endpoints == nil {
		return "", errors.New("an embeddings service has not been configured")
	}
	return c.Endpoints.Get(string(repo))
}

func (c *Client) httpPost(
	ctx context.Context,
	method string,
	repo api.RepoName,
	payload any,
) (resp *http.Response, err error) {
	url, err := c.url(repo)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	req, err := http.NewRequest("POST", url+method, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	return c.HTTPClient.Do(req)
}

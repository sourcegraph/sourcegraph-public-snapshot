package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func defaultEndpoints() *endpoint.Map {
	return endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
		return conns.Embeddings
	})
}

// TODO: Add gRPC support to the embeddings server.
var defaultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternalClientFactory("embeddings").Doer()
	if err != nil {
		panic(err)
	}
	return d
}()

func NewDefaultClient() Client {
	return NewClient(defaultEndpoints(), defaultDoer)
}

func NewClient(endpoints *endpoint.Map, doer httpcli.Doer) Client {
	return &client{
		Endpoints:  endpoints,
		HTTPClient: doer,
	}
}

type Client interface {
	Search(context.Context, EmbeddingsSearchParameters) (*EmbeddingCombinedSearchResults, error)
}

type client struct {
	// Endpoints to embeddings service.
	Endpoints *endpoint.Map

	// HTTP client to use
	HTTPClient httpcli.Doer
}

type EmbeddingsSearchParameters struct {
	RepoNames        []api.RepoName `json:"repoNames"`
	RepoIDs          []api.RepoID   `json:"repoIDs"`
	Query            string         `json:"query"`
	CodeResultsCount int            `json:"codeResultsCount"`
	TextResultsCount int            `json:"textResultsCount"`

	UseDocumentRanks bool `json:"useDocumentRanks"`
}

func (p *EmbeddingsSearchParameters) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("numRepos", len(p.RepoNames)),
		attribute.String("query", p.Query),
		attribute.Int("codeResultsCount", p.CodeResultsCount),
		attribute.Int("textResultsCount", p.TextResultsCount),
		attribute.Bool("useDocumentRanks", p.UseDocumentRanks),
	}
}

func (c *client) Search(ctx context.Context, args EmbeddingsSearchParameters) (*EmbeddingCombinedSearchResults, error) {
	partitions, err := c.partition(args.RepoNames, args.RepoIDs)
	if err != nil {
		return nil, err
	}

	p := pool.NewWithResults[*EmbeddingCombinedSearchResults]().WithContext(ctx)

	for endpoint, partition := range partitions {
		endpoint := endpoint

		// make a copy for this request
		args := args
		args.RepoNames = partition.repoNames
		args.RepoIDs = partition.repoIDs

		p.Go(func(ctx context.Context) (*EmbeddingCombinedSearchResults, error) {
			return c.searchPartition(ctx, endpoint, args)
		})
	}

	allResults, err := p.Wait()
	if err != nil {
		return nil, err
	}

	var combinedResult EmbeddingCombinedSearchResults
	for _, result := range allResults {
		combinedResult.CodeResults.MergeTruncate(result.CodeResults, args.CodeResultsCount)
		combinedResult.TextResults.MergeTruncate(result.TextResults, args.TextResultsCount)
	}

	return &combinedResult, nil
}

func (c *client) searchPartition(ctx context.Context, endpoint string, args EmbeddingsSearchParameters) (*EmbeddingCombinedSearchResults, error) {
	resp, err := c.httpPost(ctx, "search", endpoint, args)
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

	var response EmbeddingCombinedSearchResults
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type repoPartition struct {
	repoNames []api.RepoName
	repoIDs   []api.RepoID
}

// returns a partition of the input repos by the endpoint their requests should be routed to
func (c *client) partition(repos []api.RepoName, repoIDs []api.RepoID) (map[string]repoPartition, error) {
	if c.Endpoints == nil {
		return nil, errors.New("an embeddings service has not been configured")
	}

	repoStrings := make([]string, len(repos))
	for i, repo := range repos {
		repoStrings[i] = string(repo)
	}

	endpoints, err := c.Endpoints.GetMany(repoStrings...)
	if err != nil {
		return nil, err
	}

	res := make(map[string]repoPartition)
	for i, endpoint := range endpoints {
		res[endpoint] = repoPartition{
			repoNames: append(res[endpoint].repoNames, repos[i]),
			repoIDs:   append(res[endpoint].repoIDs, repoIDs[i]),
		}
	}
	return res, nil
}

func (c *client) httpPost(
	ctx context.Context,
	method string,
	url string,
	payload any,
) (resp *http.Response, err error) {
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

package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

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

var defaultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternalClientFactory("embeddings").Doer()
	if err != nil {
		panic(err)
	}
	return d
}()

func NewClient() *Client {
	return &Client{
		Endpoints:  defaultEndpoints(),
		HTTPClient: defaultDoer,
	}
}

type Client struct {
	// Endpoints to embeddings service.
	Endpoints *endpoint.Map

	// HTTP client to use
	HTTPClient httpcli.Doer
}

type EmbeddingsSearchParameters struct {
	RepoName         api.RepoName `json:"repoName"`
	RepoID           api.RepoID   `json:"repoID"`
	Query            string       `json:"query"`
	CodeResultsCount int          `json:"codeResultsCount"`
	TextResultsCount int          `json:"textResultsCount"`

	UseDocumentRanks bool `json:"useDocumentRanks"`
	// If set to "True", EmbeddingSearchResult.Debug will contain useful information about scoring.
	Debug bool `json:"debug"`
}

type IsContextRequiredForChatQueryParameters struct {
	Query string `json:"query"`
}

type IsContextRequiredForChatQueryResult struct {
	IsRequired bool `json:"isRequired"`
}

func (c *Client) Search(ctx context.Context, args EmbeddingsSearchParameters) (*EmbeddingSearchResults, error) {
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

	var response EmbeddingSearchResults
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
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

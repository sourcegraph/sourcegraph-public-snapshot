//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide access to the headers
package gerrit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Client access a Gerrit via the REST API.
type Client struct {
	// HTTP Client used to communicate with the API
	httpClient httpcli.Doer

	// Config is the code host connection config for this client
	Config *schema.GerritConnection

	// URL is the base URL of Gerrit.
	URL *url.URL

	// RateLimit is the self-imposed rate limiter (since Gerrit does not have a concept
	// of rate limiting in HTTP response headers).
	rateLimit *ratelimit.InstrumentedLimiter
}

// NewClient returns an authenticated Gerrit API client with
// the provided configuration. If a nil httpClient is provided, http.DefaultClient
// will be used.
func NewClient(urn string, config *schema.GerritConnection, httpClient httpcli.Doer) (*Client, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = httpcli.ExternalDoer
	}

	return &Client{
		httpClient: httpClient,
		Config:     config,
		URL:        u,
		rateLimit:  ratelimit.DefaultRegistry.Get(urn),
	}, nil
}

type ListAccountsResponse []Account

func (c *Client) ListAccountsByEmail(ctx context.Context, email string) (ListAccountsResponse, error) {
	qsAccounts := make(url.Values)
	qsAccounts.Set("q", fmt.Sprintf("email:%s", email)) // TODO: what query should we run?
	return c.listAccounts(ctx, qsAccounts)
}

func (c *Client) ListAccountsByUsername(ctx context.Context, username string) (ListAccountsResponse, error) {
	qsAccounts := make(url.Values)
	qsAccounts.Set("q", fmt.Sprintf("username:%s", username)) // TODO: what query should we run?
	return c.listAccounts(ctx, qsAccounts)
}

func (c *Client) listAccounts(ctx context.Context, qsAccounts url.Values) (ListAccountsResponse, error) {
	qsAccounts.Set("o", "details")

	urlPath := "a/accounts/"

	uAllProjects := url.URL{Path: urlPath, RawQuery: qsAccounts.Encode()}

	reqAllAccounts, err := http.NewRequest("GET", uAllProjects.String(), nil)

	if err != nil {
		return nil, err
	}
	respAllAccts := ListAccountsResponse{}
	if _, err = c.do(ctx, reqAllAccounts, &respAllAccts); err != nil {
		return respAllAccts, err
	}
	return respAllAccts, nil
}

// ListProjectsArgs defines options to be set on ListProjects method calls.
type ListProjectsArgs struct {
	Cursor *Pagination
}

// ListProjectsResponse defines a response struct returned from ListProjects method calls.
type ListProjectsResponse map[string]*Project

func (c *Client) ListProjects(ctx context.Context, opts ListProjectsArgs) (projects *ListProjectsResponse, nextPage bool, err error) {

	// Unfortunately Gerrit APIs are quite limited and don't support pagination well.
	// Currently, if you want to only get CODE projects and know if there is another page
	// to query for, the only way to do that is to query twice and compare the results.
	qsAllProjects := make(url.Values)
	qsCodeProjects := make(url.Values)

	if opts.Cursor == nil {
		opts.Cursor = &Pagination{PerPage: 100, Page: 1}
	}

	// Number of results to return.
	qsAllProjects.Set("n", fmt.Sprintf("%d", opts.Cursor.PerPage))
	qsCodeProjects.Set("n", fmt.Sprintf("%d", opts.Cursor.PerPage))

	// Skip the first S projects.
	qsAllProjects.Set("S", fmt.Sprintf("%d", (opts.Cursor.Page-1)*opts.Cursor.PerPage))
	qsCodeProjects.Set("S", fmt.Sprintf("%d", (opts.Cursor.Page-1)*opts.Cursor.PerPage))

	// Set the desired project type to CODE (ALL/CODE/PERMISSIONS).
	qsCodeProjects.Set("type", "CODE")

	urlPath := "a/projects/"

	uAllProjects := url.URL{Path: urlPath, RawQuery: qsAllProjects.Encode()}

	reqAllProjects, err := http.NewRequest("GET", uAllProjects.String(), nil)
	if err != nil {
		return nil, false, err
	}

	var respAllProjects ListProjectsResponse
	if _, err = c.do(ctx, reqAllProjects, &respAllProjects); err != nil {
		return nil, false, err
	}

	uCodeProjects := url.URL{Path: urlPath, RawQuery: qsCodeProjects.Encode()}

	reqCodeProjects, err := http.NewRequest("GET", uCodeProjects.String(), nil)
	if err != nil {
		return nil, false, err
	}

	var respCodeProjects ListProjectsResponse
	if _, err = c.do(ctx, reqCodeProjects, &respCodeProjects); err != nil {
		return nil, false, err
	}

	// If the amount of Projects we get back from AllProjects is greater than or equal to
	// the amount we asked for in a page, then there is another page.
	nextPage = len(respAllProjects) >= opts.Cursor.PerPage

	return &respCodeProjects, nextPage, nil
}

// nolint:unparam
func (c *Client) do(ctx context.Context, req *http.Request, result any) (*http.Response, error) {
	req.URL = c.URL.ResolveReference(req.URL)

	// Add Basic Auth headers for authenticated requests.
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Config.Username+":"+c.Config.Password)))

	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, &httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		}
	}

	// The first 4 characters of the Gerrit API responses need to be stripped, see: https://gerrit-review.googlesource.com/Documentation/rest-api.html#output .
	if len(bs) < 4 {
		return nil, &httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		}
	}
	return resp, json.Unmarshal(bs[4:], result)
}

type Account struct {
	ID          int32  `json:"_account_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
}

type Project struct {
	Description string            `json:"description"`
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Parent      string            `json:"parent"`
	State       string            `json:"state"`
	Branches    map[string]string `json:"branches"`
	Labels      map[string]Label  `json:"labels"`
}

type Label struct {
	Values       map[string]string `json:"values"`
	DefaultValue string            `json:"default_value"`
}

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Gerrit API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

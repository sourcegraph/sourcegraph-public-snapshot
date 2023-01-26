//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide access to the headers
package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Client used to access an AzureDevOps code host via the REST API.
type Client struct {
	// HTTP Client used to communicate with the API.
	httpClient httpcli.Doer

	// Config is the code host connection config for this client.
	Config *schema.AzureDevOpsConnection

	// URL is the base URL of AzureDevOps.
	URL *url.URL

	// RateLimit is the self-imposed rate limiter (since AzureDevOps does not have a concept
	// of rate limiting in HTTP response headers).
	rateLimit *ratelimit.InstrumentedLimiter
	auth      auth.Authenticator
}

// NewClient returns an authenticated AzureDevOps API client with
// the provided configuration. If a nil httpClient is provided, http.DefaultClient
// will be used.
func NewClient(urn string, config *schema.AzureDevOpsConnection, httpClient httpcli.Doer) (*Client, error) {
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
		auth: &auth.BasicAuth{
			Username: config.Username,
			Password: config.Token,
		},
	}, nil
}

// ListRepositoriesByProjectOrOrgArgs defines options to be set on the ListRepositories methods' calls.
type ListRepositoriesByProjectOrOrgArgs struct {
	// Should be in the form of 'org/project' for projects and 'org' for orgs.
	ProjectOrOrgName string
}

func (c *Client) ListRepositoriesByProjectOrOrg(ctx context.Context, opts ListRepositoriesByProjectOrOrgArgs) ([]Repository, error) {
	qs := make(url.Values)

	// TODO: @varsanojidan look into which API version/s we want to support.
	qs.Set("api-version", "7.0")

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/_apis/git/repositories", opts.ProjectOrOrgName), RawQuery: qs.Encode()}

	req, err := http.NewRequest("GET", urlRepositoriesByProjects.String(), nil)
	if err != nil {
		return nil, err
	}

	var repos ListRepositoriesResponse
	if _, err = c.do(ctx, req, &repos); err != nil {
		return nil, err
	}

	return repos.Value, nil
}

//nolint:unparam // http.Response is never used, but it makes sense API wise.
func (c *Client) do(ctx context.Context, req *http.Request, result any) (*http.Response, error) {
	req.URL = c.URL.ResolveReference(req.URL)

	// Add Basic Auth headers for authenticated requests.
	c.auth.Authenticate(req)

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

	return resp, json.Unmarshal(bs, result)
}

// WithAuthenticator returns a new Client that uses the same configuration,
// HTTPClient, and RateLimiter as the current Client, except authenticated with
// the given authenticator instance.
//
// Note that using an unsupported Authenticator implementation may result in
// unexpected behaviour, or (more likely) errors. At present, only BasicAuth is
// supported.
func (c *Client) WithAuthenticator(a auth.Authenticator) (*Client, error) {
	if _, ok := a.(*auth.BasicAuth); !ok {
		return nil, errors.Errorf("authenticator type unsupported for Azure DevOps clients: %s", a)
	}

	return &Client{
		httpClient: c.httpClient,
		URL:        c.URL,
		auth:       a,
		rateLimit:  c.rateLimit,
	}, nil
}

type ListRepositoriesResponse struct {
	Value []Repository `json:"value"`
	Count int          `json:"count"`
}

type Repository struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CloneURL   string  `json:"remoteURL"`
	APIURL     string  `json:"url"`
	SSHURL     string  `json:"sshUrl"`
	WebURL     string  `json:"webUrl"`
	IsDisabled bool    `json:"isDisabled"`
	Project    Project `json:"project"`
}

type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	Revision   int    `json:"revision"`
	Visibility string `json:"visibility"`
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Azure DevOps API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

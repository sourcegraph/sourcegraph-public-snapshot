//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide access to the headers
package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	azureDevOpsServicesURL = "https://dev.azure.com/"
	// TODO: @varsanojidan look into which API version/s we want to support.
	apiVersion              = "7.0"
	continuationTokenHeader = "x-ms-continuationtoken"
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
func NewClient(urn string, url *url.URL, auth auth.Authenticator, httpClient httpcli.Doer) (*Client, error) {
	if httpClient == nil {
		httpClient = httpcli.ExternalDoer
	}

	return &Client{
		httpClient: httpClient,
		// Config:     config,
		URL:       url,
		rateLimit: ratelimit.DefaultRegistry.Get(urn),
		auth:      auth,
	}, nil
}

// TODO: Code organisation

type ClientProvider struct {
	// The URN of the external service that the client is derived from.
	// TODO: Where is this used?
	urn string

	// baseURL is the base URL of GitLab; e.g., https://gitlab.com or https://gitlab.example.com
	// TODO: Do I need this?
	baseURL *url.URL

	// httpClient is the underlying the HTTP client to use.
	httpClient httpcli.Doer

	clients   map[string]*Client
	clientsMu sync.Mutex
}

// GetOAuthClient returns a client authenticated by the OAuth token.
func (p *ClientProvider) GetOAuthClient(oauthToken string) *Client {
	if oauthToken == "" {
		return p.getClient(nil)
	}
	return p.getClient(&auth.OAuthBearerToken{Token: oauthToken})
}

// GetClient returns an unauthenticated client.
func (p *ClientProvider) GetClient() *Client {
	// TODO: Maybe inline.
	// return p.getClient(nil)

	// FIXME
	return nil
}

// TODO: Move inline if only place used is above.
func (p *ClientProvider) getClient(a auth.Authenticator) *Client {
	// FIXME
	return nil
	// p.clientsMu.Lock()
	// defer p.clientsMu.Unlock()

	// key := "<nil>"
	// if a != nil {
	// 	key = a.Hash()
	// }
	// if c, ok := p.clients[key]; ok {
	// 	return c
	// }

	// c, err := NewClient("AzureDevOpsOAuth", a)
	// if err != nil {
	// 	// TODO: Do not panic.
	// 	panic(err)
	// }

	// p.clients[key] = c
	// return c
}

// func (p *ClientProvider) NewClient(a auth.Authenticator) *Client {
// 	// Cache for GitLab project metadata.
// 	// var cacheTTL time.Duration
// 	// if isGitLabDotComURL(p.baseURL) && a == nil {
// 	// 	cacheTTL = 10 * time.Minute // cache for longer when unauthenticated
// 	// } else {
// 	// 	cacheTTL = 30 * time.Second
// 	// }
// 	// key := "gl_proj:"
// 	// var tokenHash string
// 	// if a != nil {
// 	// 	tokenHash = a.Hash()
// 	// 	key += tokenHash
// 	// }e
// 	// projCache := rcache.NewWithTTL(key, int(cacheTTL/time.Second))

// 	// rl := ratelimit.DefaultRegistry.Get(p.urn)
// 	// rlm := ratelimit.DefaultMonitorRegistry.GetOrSet(p.baseURL.String(), tokenHash, "rest", &ratelimit.Monitor{})

// 	return &Client{
// 		httpClient: httpcli.ExternalDoer,
// 		// Config:     config,
// 		URL:       u,
// 		rateLimit: ratelimit.DefaultRegistry.Get(urn),
// 		auth:      a,
// 		// 	Username: config.Username,
// 		// 	Password: config.Token,
// 		// },
// 	}
// }

func NewClientProvider(urn string, baseURL *url.URL, cli httpcli.Doer) *ClientProvider {
	if cli == nil {
		cli = httpcli.ExternalDoer
	}

	return &ClientProvider{
		urn: urn,
		// TODO: What do we do here?
		// baseURL:    baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, "api/v4") + "/"}),
		httpClient: cli,
		clients:    make(map[string]*Client),
	}
}

// do performs the specified request, returning any errors and a continuationToken used for pagination (if the API supports it).
//
//nolint:unparam // http.Response is never used, but it makes sense API wise.
func (c *Client) do(ctx context.Context, req *http.Request, urlOverride string, result any) (continuationToken string, err error) {
	u := c.URL
	if urlOverride != "" {
		u, err = url.Parse(urlOverride)
		if err != nil {
			return "", err
		}
	}

	queryParams := req.URL.Query()
	queryParams.Set("api-version", apiVersion)
	req.URL.RawQuery = queryParams.Encode()
	req.URL = u.ResolveReference(req.URL)
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add Basic Auth headers for authenticated requests.
	c.auth.Authenticate(req)

	if err := c.rateLimit.Wait(ctx); err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", &httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		}
	}

	return resp.Header.Get(continuationTokenHeader), json.Unmarshal(bs, result)
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

// IsAzureDevOpsServices returns true if the client is configured to Azure DevOps
// Services (https://dev.azure.com
func (c *Client) IsAzureDevOpsServices() bool {
	return c.URL.String() == azureDevOpsServicesURL
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Azure DevOps API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide access to the headers
package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/goware/urlx"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

const (
	AZURE_DEV_OPS_API_URL = "https://dev.azure.com/"
	// TODO: @varsanojidan look into which API version/s we want to support.
	apiVersion              = "7.0"
	continuationTokenHeader = "x-ms-continuationtoken"
)

// Client used to access an AzureDevOps code host via the REST API.
type Client struct {
	// HTTP Client used to communicate with the API.
	httpClient httpcli.Doer

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
func NewClient(urn string, url string, auth auth.Authenticator, httpClient httpcli.Doer) (*Client, error) {
	u, err := urlx.Parse(url)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = httpcli.ExternalDoer
	}

	return &Client{
		httpClient: httpClient,
		URL:        u,
		rateLimit:  ratelimit.DefaultRegistry.Get(urn),
		auth:       auth,
	}, nil
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

	// Add authentication headers for authenticated requests.
	c.auth.Authenticate(req)

	if err := c.rateLimit.Wait(ctx); err != nil {
		return "", err
	}

	logger := log.Scoped("azuredevops.Client", "azuredevops Client logger")
	resp, err := oauthutil.DoRequest(ctx, logger, c.httpClient, req, c.auth)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", &HTTPError{
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
	return c.URL.String() == AZURE_DEV_OPS_API_URL
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("Azure DevOps API HTTP error: code=%d url=%q", e.StatusCode, e.URL)
}

func GetOAuthContext(refreshToken string) (*oauthutil.OAuthContext, error) {
	for _, authProvider := range conf.SiteConfig().AuthProviders {
		if authProvider.AzureDevOps != nil {
			authURL, err := url.JoinPath(VISUAL_STUDIO_APP_URL, "oauth2/authorize")
			if err != nil {
				continue
			}
			tokenURL, err := url.JoinPath(VISUAL_STUDIO_APP_URL, "oauth2/token")
			if err != nil {
				continue
			}

			redirectURL, err := GetRedirectURL()
			if err != nil {
				return nil, err
			}

			p := authProvider.AzureDevOps
			return &oauthutil.OAuthContext{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  authURL,
					TokenURL: tokenURL,
				},
				AdditionalArgs: map[string]string{
					"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
					"client_assertion":      url.QueryEscape(p.ClientSecret),
					"grant_type":            "refresh_token",
					"assertion":             url.QueryEscape(refreshToken),
					"redirect_uri":          redirectURL.String(),
				},
			}, nil
		}
	}

	return nil, errors.New("No authprovider configured for AzureDevOps, check site configuraiton.")
}

func GetRedirectURL() (*url.URL, error) {
	externalURL, err := url.Parse(conf.SiteConfig().ExternalURL)
	if err != nil {
		return nil, errors.New("Could not parse `externalURL`, which is needed to determine the OAuth callback URL.")
	}

	callbackURL := *externalURL
	callbackURL.Path = "/.auth/azuredevops/callback"
	return &callbackURL, nil
}

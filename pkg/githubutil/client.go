package githubutil

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"

	"github.com/sourcegraph/go-github/github"
	"github.com/sourcegraph/httpcache"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Config specifies configuration options for a GitHub API client used
// by Sourcegraph code.
type Config struct {
	BaseURL       *url.URL          // base URL of GitHub API; e.g., https://api.github.com (or ghcompat URL)
	OAuth         oauth2.Config     // OAuth config
	AppdashSpanID appdash.SpanID    // if nonzero, trace API calls and associate them with this span
	Cache         httpcache.Cache   // if set, caches HTTP responses (namespaced per-token for authed client)
	CacheControl  string            // cache-control header to set on all client requests, if non-empty
	Transport     http.RoundTripper // base HTTP transport (if nil, uses http.DefaultTransport)
}

// UnauthedClient is a GitHub API client using the config's OAuth2
// client ID and secret, but not using any specific user's access
// token. It enables a higher rate limit (5000 per hour instead of 60
// per hour, as of Nov 2014).
func (c *Config) UnauthedClient() *github.Client {
	var t http.RoundTripper = c.baseTransport()
	if c.Cache != nil {
		t = &httpcache.Transport{
			Cache:               c.Cache,
			Transport:           t,
			MarkCachedResponses: true,
		}
	}
	t = NewGitHubCacheControlTransport(c.CacheControl, t)
	if c.OAuth.ClientID != "" {
		t = &github.UnauthenticatedRateLimitedTransport{
			ClientID:     c.OAuth.ClientID,
			ClientSecret: c.OAuth.ClientSecret,
			Transport:    t,
		}
	}

	return c.client(&http.Client{Transport: t})
}

// AuthedTransport returns a GitHub HTTP transport using a user's
// OAuth2 access token. All actions taken by clients using this
// transport will use the full granted permissions of the token's
// user. It uses a HTTP cache transport whose storage keys are
// namespaced by token so that private information does not leak
// across users.
func (c *Config) AuthedClient(token string) *github.Client {
	var t http.RoundTripper = c.baseTransport()

	tokHash := sha256.Sum256([]byte(token))
	if c.Cache != nil {
		t = &httpcache.Transport{
			Cache: namespacedCache{
				namespace: base64.URLEncoding.EncodeToString(tokHash[:]),
				Cache:     c.Cache,
			},
			Transport:           t,
			MarkCachedResponses: true,
		}
	}

	t = NewGitHubCacheControlTransport(c.CacheControl, t)
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: t})
	return c.client(c.OAuth.Client(ctx, &oauth2.Token{AccessToken: token}))
}

// ApplicationAuthedClient returns a GitHub API client that sends the
// OAuth2 application's client credentials in HTTP basic auth. It is
// necessary to do this for some API endpoints, such as the "revoke an
// authorization for an application" endpoint.
//
// This is different from UnauthedClient, which uses the application's
// client ID and secret, but passes them in the URL, not the HTTP
// Authorization header. GitHub treats those two differently.
func (c *Config) ApplicationAuthedClient() *github.Client {
	// No need for caching; this is a rarely used client and is only
	// used uncached for POST/DELETE operations anyway.

	var t http.RoundTripper = c.baseTransport()

	t = &github.BasicAuthTransport{
		Username:  c.OAuth.ClientID,
		Password:  c.OAuth.ClientSecret,
		Transport: t,
	}

	return c.client(&http.Client{Transport: t})
}

// client creates a new GitHub API client from the transport.
func (c *Config) client(httpClient *http.Client) *github.Client {
	{
		// Avoid modifying httpClient.
		tmp := *httpClient
		tmp.Transport = c.applyAppdash(tmp.Transport)
		httpClient = &tmp
	}

	g := github.NewClient(httpClient)
	if c.BaseURL != nil {
		g.BaseURL = c.BaseURL
	}
	return g
}

func (c *Config) baseTransport() http.RoundTripper {
	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	if githubcli.Config.Disable {
		transport = disabledTransport{}
	}

	// Instrument metrics before the RetryTransport to get a better
	// understanding of our responses from GitHub
	transport = &metricsTransport{Transport: transport}

	// Retry GitHub API requests (sometimes the connection is dropped,
	// and we don't want to fail the whole request tree because of 1
	// ephemeral error out of possibly tens of GitHub API requests).
	transport = &httputil.RetryTransport{
		Retries:   2,
		Delay:     time.Millisecond * 100,
		Transport: transport,
	}

	return transport
}

// applyAppdash wraps t in an appdash transport if appdash is
// enabled in the config. Otherwise, it returns t as-is.
//
// This is not part of baseTransport because we want to insert
// appdash before the UnauthenticatedRateLimitedTransport, or else
// our client secret would leak all over appdash.
func (c *Config) applyAppdash(t http.RoundTripper) http.RoundTripper {
	if c.AppdashSpanID.Trace == 0 {
		return t
	}
	return &httptrace.Transport{
		Recorder:  traceutil.NewRecorder(c.AppdashSpanID, traceutil.DefaultCollector),
		Transport: t,
		SetName:   true,
	}
}

type disabledTransport struct{}

func (t disabledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, errors.New("http: github communication disabled")
}

// Default is the default configuration for the GitHub API client, with auth and token URLs for github.com and client ID/secret values taken from the environment.
var Default = &Config{
	BaseURL: &url.URL{Scheme: "https", Host: "api.github.com"},
	OAuth: oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
			// RedirectURL is not known at init time, so it's not set here.
		},
	},
	Cache: httputil.Cache,
}

func NewTestClientServer() (client *github.Client, config *Config, mux *http.ServeMux) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		zero := 0
		result := github.RepositoriesSearchResult{
			Total:        &zero,
			Repositories: []github.Repository{},
		}
		w.Header().Set("content-type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(result)
	})

	var err error
	client = github.NewClient(nil)
	client.BaseURL, err = url.Parse(server.URL)
	if err != nil {
		log.Panicf("Could not create new test client: %v", err)
	}
	config = &Config{BaseURL: client.BaseURL}
	return
}

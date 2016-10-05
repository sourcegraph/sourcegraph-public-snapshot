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
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"

	"context"

	"github.com/golang/groupcache/lru"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	"github.com/sourcegraph/httpcache"
	"golang.org/x/oauth2"
)

func init() {
	prometheus.MustRegister(githubUnauthedConcurrent)
	prometheus.MustRegister(reposGitHubHTTPCacheCounter)
}

// Config specifies configuration options for a GitHub API client used
// by Sourcegraph code.
type Config struct {
	BaseURL          *url.URL          // base URL of GitHub API; e.g., https://api.github.com (or ghcompat URL)
	OAuth            oauth2.Config     // OAuth config
	Context          context.Context   // the context for requests to GitHub
	Cache            httpcache.Cache   // if set, caches HTTP responses (namespaced per-token for authed client)
	CacheControl     string            // cache-control header to set on all client requests, if non-empty
	Transport        http.RoundTripper // base HTTP transport (if nil, uses http.DefaultTransport)
	UnauthedThrottle sync.Locker       // limits number of concurrent requests to GitHub for all unauthenticated clients
	UserThrottles    *ThrottleCache    // limits number of concurrent requests to GitHub per user token
}

// ThrottleCache stores throttles for a given user so they can be reused across
// different requests.
type ThrottleCache struct {
	sync.Mutex
	Throttles *lru.Cache // GitHub access token -> *sync.Mutex
}

var githubUnauthedConcurrent = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "github",
	Name:      "unauthed_concurrent",
	Help:      "Number of unauthed concurrent calls to GitHub's Repo API.",
})

var reposGitHubHTTPCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "repos",
	Name:      "github_api_cache_hit",
	Help:      "Counts cache hits and misses for the github API HTTP cache.",
}, []string{"type"})

// cacheWithMetrics tracks the number of cache hits and misses returned from an
// httpcache.Cache in prometheus.
type cacheWithMetrics struct {
	cache   httpcache.Cache
	counter *prometheus.CounterVec
}

func (c *cacheWithMetrics) Get(key string) ([]byte, bool) {
	resp, ok := c.cache.Get(key)
	if ok {
		c.counter.WithLabelValues("hit").Inc()
	} else {
		c.counter.WithLabelValues("miss").Inc()
	}
	return resp, ok
}

func (c *cacheWithMetrics) Set(key string, resp []byte) {
	c.cache.Set(key, resp)
}

func (c *cacheWithMetrics) Delete(key string) {
	c.cache.Delete(key)
}

// UnauthedClient is a GitHub API client using the config's OAuth2
// client ID and secret, but not using any specific user's access
// token. It enables a higher rate limit (5000 per hour instead of 60
// per hour, as of Nov 2014).
func (c *Config) UnauthedClient() *github.Client {
	var t http.RoundTripper = baseTransport(c.Transport)
	t = &throttledTransport{
		Transport: t,
		Throttle:  c.UnauthedThrottle,
	}
	if c.Cache != nil {
		t = &httpcache.Transport{
			Cache:               c.Cache,
			Transport:           t,
			MarkCachedResponses: true,
		}
	}
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
	var t http.RoundTripper = baseTransport(c.Transport)
	if c.UserThrottles != nil {
		t = &throttledTransport{
			Transport: t,
			Throttle:  c.UserThrottles.Get(token),
		}
	}

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

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: t})
	return c.client(c.OAuth.Client(ctx, &oauth2.Token{AccessToken: token}))
}

func (t *ThrottleCache) Get(token string) *sync.Mutex {
	t.Lock()
	defer t.Unlock()
	v, ok := t.Throttles.Get(token)
	if ok {
		return v.(*sync.Mutex)
	}
	m := &sync.Mutex{}
	t.Throttles.Add(token, m)
	return m
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

	var t http.RoundTripper = &throttledTransport{
		Transport: baseTransport(c.Transport),
		Throttle:  c.UnauthedThrottle,
	}

	t = &github.BasicAuthTransport{
		Username:  c.OAuth.ClientID,
		Password:  c.OAuth.ClientSecret,
		Transport: t,
	}

	return c.client(&http.Client{Transport: t})
}

// client creates a new GitHub API client from the transport.
func (c *Config) client(httpClient *http.Client) *github.Client {
	if c.Context == nil {
		c.Context = context.Background()
	}
	{
		// Avoid modifying httpClient.
		tmp := *httpClient
		tmp.Transport = &tracingTransport{tmp.Transport, c.Context}
		httpClient = &tmp
	}

	g := github.NewClient(httpClient)
	if c.BaseURL != nil {
		g.BaseURL = c.BaseURL
	}
	return g
}

func baseTransport(transport http.RoundTripper) http.RoundTripper {
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

type tracingTransport struct {
	t   http.RoundTripper
	ctx context.Context
}

func (t *tracingTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	span, ctx := opentracing.StartSpanFromContext(t.ctx, "GitHub")
	span.SetTag("URL", req.URL.String())
	defer func() {
		if err != nil {
			span.SetTag("error", err.Error())
		}
		if resp != nil {
			span.SetTag("status", resp.Status)
		}
		span.Finish()
	}()

	resp, err = t.t.RoundTrip(req.WithContext(ctx))
	return
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
	Cache: &cacheWithMetrics{
		cache:   httputil.Cache,
		counter: reposGitHubHTTPCacheCounter,
	},
	UnauthedThrottle: newGaugedMutex(githubUnauthedConcurrent),
	UserThrottles: &ThrottleCache{
		Throttles: lru.New(1000),
	},
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

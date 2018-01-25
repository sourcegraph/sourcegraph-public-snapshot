package github

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context/ctxhttp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"

	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	gitHubDisable, _ = strconv.ParseBool(env.Get("SRC_GITHUB_DISABLE", "false", "disables communication with GitHub instances. Used to test GitHub service degredation"))
	githubProxyURL   = func() *url.URL {
		url, err := url.Parse(env.Get("GITHUB_BASE_URL", "http://github-proxy", "base URL for GitHub.com API (used for github-proxy)"))
		if err != nil {
			log.Fatal("Error parsing GITHUB_BASE_URL:", err)
		}
		return url
	}()
)

// Client is a GitHub API client.
type Client struct {
	baseURL    *url.URL     // base URL of GitHub API; e.g., https://api.github.com
	token      string       // a personal access token to authenticate requests, if set
	httpClient *http.Client // the HTTP client to use

	repoCache *rcache.Cache

	mu                 sync.Mutex
	rateLimitKnown     bool
	rateLimit          int       // last X-RateLimit-Limit HTTP response header value
	rateLimitRemaining int       // last X-RateLimit-Remaining HTTP response header value
	rateLimitReset     time.Time // last X-RateLimit-Remaining HTTP response header value
}

// NewClient creates a new GitHub API client with an optional personal access token to authenticate requests.
func NewClient(baseURL *url.URL, token string, transport http.RoundTripper) *Client {
	if gitHubDisable {
		transport = disabledTransport{}
	}
	if transport == nil {
		transport = http.DefaultTransport
	}
	transport = &metricsTransport{Transport: transport}

	var cacheTTL time.Duration
	if hostname := strings.ToLower(baseURL.Hostname()); hostname == "api.github.com" || hostname == "github.com" || hostname == "www.github.com" {
		cacheTTL = 10 * time.Minute
		// For GitHub.com API requests, use github-proxy (which adds our OAuth2 client ID/secret to get a much higher
		// rate limit).
		baseURL = githubProxyURL
	} else {
		// GitHub Enterprise
		cacheTTL = 30 * time.Second
		if baseURL.Path == "" || baseURL.Path == "/" {
			baseURL = baseURL.ResolveReference(&url.URL{Path: "/api"})
		}
	}

	// Cache for repository metadata.
	key := sha256.Sum256([]byte(token + ":" + baseURL.String()))
	repoCache := rcache.NewWithTTL("gh_repo:"+base64.URLEncoding.EncodeToString(key[:]), int(cacheTTL/time.Second))

	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Transport: transport},
		repoCache:  repoCache,
	}
}

// RateLimit reports the client's GitHub rate limit (as of the last API response it received).
func (c *Client) RateLimit() (remaining int, reset time.Duration, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.rateLimitKnown {
		return 0, 0, false
	}
	return c.rateLimitRemaining, c.rateLimitReset.Sub(time.Now()), false
}

// RecommendedRateLimitWaitForBackgroundOp returns the recommended wait time before performing a periodic
// background operation with the given rate limit cost. It takes the rate limit information from the last API
// request into account.
//
// For example, suppose the rate limit resets to 5,000 points in 30 minutes and currently 1,500 points remain. You
// want to perform a cost-500 operation. Only 4 more cost-500 operations are allowed in the next 30 minutes (per
// the rate limit), so a recommended wait time would be N.
//
//                          -500         -500         -500
//         Now   |------------*------------*------------*------------| 30 min from now
//   Remaining  1500         1000         500           0           5000 (reset)
//
// Assuming no other operations are being performed (that count against the rate limit), the recommended wait would
// be 7.5 minutes (30 minutes / 4), so that the operations are evenly spaced out.
//
// A small constant additional wait is added to account for other simultaneous operations and clock
// out-of-synchronization with GitHub.
//
// See https://developer.github.com/v4/guides/resource-limitations/#rate-limit.
func (c *Client) RecommendedRateLimitWaitForBackgroundOp(cost int) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.rateLimitKnown {
		return 0
	}

	// If our rate limit info is out of date, assume it was reset.
	limitRemaining := float64(c.rateLimitRemaining)
	resetAt := c.rateLimitReset
	if time.Now().Before(c.rateLimitReset) {
		limitRemaining = float64(c.rateLimit)
		resetAt = time.Now().Add(1 * time.Hour)
	}

	// Be conservative.
	limitRemaining = float64(limitRemaining)*0.8 - 150
	timeRemaining := resetAt.Sub(time.Now()) + 3*time.Minute

	n := limitRemaining / float64(cost) // number of times this op can run before exhausting rate limit
	if n < 1 {
		n = 1 // no point in waiting beyond the reset
	}
	if n > 500 {
		return 0
	}
	if n > 250 {
		return 200 * time.Millisecond
	}
	return (timeRemaining / time.Duration(n)).Round(time.Second)
}

func (c *Client) recordRateLimitHeaders(h http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// See https://developer.github.com/v3/#rate-limiting.
	limit, err := strconv.Atoi(h.Get("X-RateLimit-Limit"))
	if err != nil {
		c.rateLimitKnown = false
		return
	}
	remaining, err := strconv.Atoi(h.Get("X-RateLimit-Remaining"))
	if err != nil {
		c.rateLimitKnown = false
		return
	}
	resetAtSeconds, err := strconv.ParseInt(h.Get("X-RateLimit-Reset"), 10, 64)
	if err != nil {
		c.rateLimitKnown = false
		return
	}
	c.rateLimitKnown = true
	c.rateLimit = limit
	c.rateLimitRemaining = remaining
	c.rateLimitReset = time.Unix(resetAtSeconds, 0)
}

func (c *Client) requestGraphQL(ctx context.Context, query string, vars map[string]interface{}, result interface{}) (err error) {
	reqBody, err := json.Marshal(struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: vars,
	})
	if err != nil {
		return err
	}
	url := c.baseURL.ResolveReference(&url.URL{Path: path.Join(c.baseURL.Path, "graphql")})
	req, err := http.NewRequest("POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.token != "" {
		req.Header.Set("Authorization", "bearer "+c.token)
	}

	var resp *http.Response

	span, ctx := opentracing.StartSpanFromContext(ctx, "GitHub")
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

	resp, err = ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.recordRateLimitHeaders(resp.Header)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP error status %d from GitHub GraphQL endpoint (%s)", resp.StatusCode, req.URL)
	}

	var respBody struct {
		Data   json.RawMessage `json:"data"`
		Errors graphqlErrors   `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}
	if len(respBody.Errors) > 0 {
		return respBody.Errors
	}

	if result != nil {
		if err := json.Unmarshal(respBody.Data, result); err != nil {
			return err
		}
	}
	return nil
}

// graphqlErrors describes the errors in a GraphQL response. It contains at least 1 element when returned by
// requestGraphQL. See https://facebook.github.io/graphql/#sec-Errors.
type graphqlErrors []struct {
	Message   string   `json:"message"`
	Type      string   `json:"type"`
	Path      []string `json:"path"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

func (e graphqlErrors) Error() string {
	return fmt.Sprintf("error in GraphQL response: %s", e[0].Message)
}

type disabledTransport struct{}

func (t disabledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, errors.New("http: github communication disabled")
}

var reposGitHubHTTPCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "repos",
	Name:      "github_api_cache_hit",
	Help:      "Counts cache hits and misses for the github API HTTP cache.",
}, []string{"type"})

func init() {
	prometheus.MustRegister(reposGitHubHTTPCacheCounter)
}

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

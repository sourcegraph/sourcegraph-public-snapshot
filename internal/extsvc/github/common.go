package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

var (
	gitHubDisable, _ = strconv.ParseBool(env.Get("SRC_GITHUB_DISABLE", "false", "disables communication with GitHub instances. Used to test GitHub service degradation"))
	githubProxyURL   = func() *url.URL {
		url, err := url.Parse(env.Get("GITHUB_BASE_URL", "http://github-proxy", "base URL for GitHub.com API (used for github-proxy)"))
		if err != nil {
			log.Fatal("Error parsing GITHUB_BASE_URL:", err)
		}
		return url
	}()

	requestCounter = metrics.NewRequestMeter("github", "Total number of requests sent to the GitHub API.")
)

func doRequest(ctx context.Context, apiURL *url.URL, auth auth.Authenticator, rateLimitMonitor *ratelimit.Monitor, httpClient httpcli.Doer, req *http.Request, result interface{}) (err error) {
	req.URL.Path = path.Join(apiURL.Path, req.URL.Path)
	req.URL = apiURL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if auth != nil {
		if err := auth.Authenticate(req); err != nil {
			return errors.Wrap(err, "authenticating request")
		}
	}

	var resp *http.Response

	span, ctx := ot.StartSpanFromContext(ctx, "GitHub")
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

	resp, err = httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	rateLimitMonitor.Update(resp.Header)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		var err APIError
		if body, readErr := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<13)); readErr != nil { // 8kb
			err.Message = fmt.Sprintf("failed to read error response from GitHub API: %v: %q", readErr, string(body))
		} else if decErr := json.Unmarshal(body, &err); decErr != nil {
			err.Message = fmt.Sprintf("failed to decode error response from GitHub API: %v: %q", decErr, string(body))
		}
		err.URL = req.URL.String()
		err.Code = resp.StatusCode
		return &err
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func canonicalizedURL(apiURL *url.URL) *url.URL {
	if urlIsGitHubDotCom(apiURL) {
		// For GitHub.com API requests, use github-proxy (which adds our OAuth2 client ID/secret to get a much higher
		// rate limit).
		return githubProxyURL
	}
	return apiURL
}

func urlIsGitHubDotCom(apiURL *url.URL) bool {
	hostname := strings.ToLower(apiURL.Hostname())
	return hostname == "api.github.com" || hostname == "github.com" || hostname == "www.github.com" || apiURL.String() == githubProxyURL.String()
}

// ErrNotFound is when the requested GitHub repository is not found.
var ErrNotFound = errors.New("GitHub repository not found")

// IsNotFound reports whether err is a GitHub API error of type NOT_FOUND, the equivalent cached
// response error, or HTTP 404.
func IsNotFound(err error) bool {
	if err == ErrNotFound || errors.Cause(err) == ErrNotFound {
		return true
	}
	if _, ok := err.(ErrPullRequestNotFound); ok {
		return true
	}
	if HTTPErrorCode(err) == http.StatusNotFound {
		return true
	}
	errs, ok := err.(graphqlErrors)
	if !ok {
		return false
	}
	for _, err := range errs {
		if err.Type == "NOT_FOUND" {
			return true
		}
	}
	return false
}

type disabledClient struct{}

func (t disabledClient) Do(r *http.Request) (*http.Response, error) {
	return nil, errors.New("http: github communication disabled")
}

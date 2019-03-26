package bitbucketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/metrics"
	"golang.org/x/time/rate"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var requestCounter = metrics.NewRequestCounter("bitbucket", "Total number of requests sent to the Bitbucket API.")

// These fields define the self-imposed Bitbucket rate limit (since Bitbucket Server does
// not have a concept of rate limiting in HTTP response headers).
//
// See https://godoc.org/golang.org/x/time/rate#Limiter for an explanation of these fields.
//
// The limits chosen here are based on the following logic: Bitbucket Cloud restricts
// "List all repositories" requests (which are a good portion of our requests) to 1,000/hr,
// and they restrict "List a user or team's repositories" requests (which are roughly equal
// to our repository lookup requests) to 1,000/hr. We peform a list repositories request
// for every 100 repositories on Bitbucket every 1m by default, so for someone with 20,000
// Bitbucket repositories we need 20,000/100 requests per minute (1200/hr) + overhead for
// repository lookup requests by users. So we use a generous 7,200/hr here until we hear
// from someone that these values do not work well for them.
const (
	rateLimitRequestsPerSecond = 2 // 120/min or 7200/hr
	RateLimitMaxBurstRequests  = 500
)

// Client access a Bitbucket Server via the REST API.
type Client struct {
	// HTTP Client used to communicate with the API
	httpClient httpcli.Doer

	// URL is the base URL of Bitbucket Server.
	URL *url.URL

	// Token is the personal access token for accessing the
	// server. https://bitbucket.example.com/plugins/servlet/access-tokens/manage
	Token string

	// The username and password credentials for accessing the server. Typically these are only
	// used when the server doesn't support personal access tokens (such as Bitbucket Server
	// version 5.4 and older). If both Token and Username/Password are specified, Token is used.
	Username, Password string

	// RateLimit is the self-imposed rate limiter (since Bitbucket does not have a concept
	// of rate limiting in HTTP response headers).
	RateLimit *rate.Limiter
}

// NewClient returns a new Bitbucket Server API client at url. If a nil
// httpClient is provided, http.DefaultClient will be used. To use API methods
// which require authentication, set the Token or Username/Password fields of
// the returned client.
func NewClient(url *url.URL, httpClient httpcli.Doer) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	httpClient = requestCounter.Doer(httpClient, categorize)

	return &Client{
		httpClient: httpClient,
		URL:        url,
		RateLimit:  rate.NewLimiter(rateLimitRequestsPerSecond, RateLimitMaxBurstRequests),
	}
}

func (c *Client) Repo(ctx context.Context, projectKey, repoSlug string) (*Repo, error) {
	u := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s", projectKey, repoSlug)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var resp Repo
	err = c.do(ctx, req, &resp)
	return &resp, err
}

func (c *Client) Repos(ctx context.Context, pageToken *PageToken) ([]*Repo, *PageToken, error) {
	u := fmt.Sprintf("rest/api/1.0/repos%s", pageToken.Query())
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	var resp struct {
		*PageToken
		Values []*Repo
	}
	err = c.do(ctx, req, &resp)
	if err != nil {
		return nil, nil, err
	}
	return resp.Values, resp.PageToken, nil
}

func (c *Client) RecentRepos(ctx context.Context, pageToken *PageToken) ([]*Repo, *PageToken, error) {
	u := fmt.Sprintf("rest/api/1.0/profile/recent/repos%s", pageToken.Query())
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	var resp struct {
		*PageToken
		Values []*Repo
	}
	err = c.do(ctx, req, &resp)
	if err != nil {
		return nil, nil, err
	}
	return resp.Values, resp.PageToken, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) error {
	req.URL = c.URL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Authenticate request, preferring token.
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else if c.Username != "" || c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Bitbucket Server"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	startWait := time.Now()
	if err := c.RateLimit.Wait(ctx); err != nil {
		return err
	}
	if d := time.Since(startWait); d > 200*time.Millisecond {
		log15.Warn("Bitbucket self-enforced API rate limit: request delayed longer than expected due to rate limit", "delay", d)
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	// If we got an error, and the context has been canceled,
	// the context's error is probably more useful.
	if err != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.WithStack(&httpError{URL: req.URL, StatusCode: resp.StatusCode})
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// categorize returns a category for an API URL. Used by metrics.
func categorize(u *url.URL) string {
	// API to URL mapping looks like this:
	//
	// 	Repo -> rest/api/1.0/profile/recent/repos%s
	// 	Repos -> rest/api/1.0/projects/%s/repos/%s
	// 	RecentRepos -> rest/api/1.0/repos%s
	//
	// We guess the category based on the fourth path component ("profile", "projects", "repos%s").
	var category string
	if parts := strings.SplitN(u.Path, "/", 3); len(parts) >= 4 {
		category = parts[3]
	}
	switch {
	case category == "profile":
		return "Repo"
	case category == "projects":
		return "Repos"
	case strings.HasPrefix(category, "repos"):
		return "RecentRepos"
	default:
		// don't return category directly as that could introduce too much dimensionality
		return "unknown"
	}
}

type PageToken struct {
	Size          int  `json:"size"`
	Limit         int  `json:"limit"`
	IsLastPage    bool `json:"isLastPage"`
	Start         int  `json:"start"`
	NextPageStart int  `json:"nextPageStart"`
}

func (t *PageToken) HasMore() bool {
	if t == nil {
		return true
	}
	return !t.IsLastPage
}

func (t *PageToken) Query() string {
	if t == nil {
		return ""
	}
	v := url.Values{}
	if t.NextPageStart != 0 {
		v.Set("start", strconv.Itoa(t.NextPageStart))
	}
	if t.Limit != 0 {
		v.Set("limit", strconv.Itoa(t.Limit))
	}
	if len(v) == 0 {
		return ""
	}
	return "?" + v.Encode()
}

type Repo struct {
	Slug          string   `json:"slug"`
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	SCMID         string   `json:"scmId"`
	State         string   `json:"state"`
	StatusMessage string   `json:"statusMessage"`
	Forkable      bool     `json:"forkable"`
	Origin        *Repo    `json:"origin"`
	Project       *Project `json:"project"`
	Public        bool     `json:"public"`
	Links         struct {
		Clone []struct {
			Href string `json:"href"`
			Name string `json:"name"`
		} `json:"clone"`
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

// IsPersonalRepository tells if the repository is a personal one.
func (r *Repo) IsPersonalRepository() bool {
	return r.Project.Type == "PERSONAL"
}

type Project struct {
	Key    string `json:"key"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
	Links  struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type httpError struct {
	StatusCode int
	URL        *url.URL
}

func (e *httpError) Error() string {
	return fmt.Sprintf("unexpected %d response from BitBucket Server REST API at %s", e.StatusCode, e.URL)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

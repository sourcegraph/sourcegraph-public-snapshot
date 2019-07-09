package bitbucketcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/metrics"
	"golang.org/x/time/rate"
	"gopkg.in/inconshreveable/log15.v2"
)

var requestCounter = metrics.NewRequestCounter("bitbucket_cloud_requests_count", "Total number of requests sent to the Bitbucket Cloud API.")

// These fields define the self-imposed Bitbucket rate limit (since Bitbucket Cloud does
// not have a concept of rate limiting in HTTP response headers).
//
// See https://godoc.org/golang.org/x/time/rate#Limiter for an explanation of these fields.
//
// The limits chosen here are based on the following logic: Bitbucket Cloud restricts
// "List all repositories" requests (which are a good portion of our requests) to 1,000/hr,
// and they restrict "List a user or team's repositories" requests (which are roughly equal
// to our repository lookup requests) to 1,000/hr. We peform a list repositories request
// for every 1000 repositories on Bitbucket every 1m by default, so for someone with 20,000
// Bitbucket repositories we need 20,000/1000 requests per minute (1200/hr) + overhead for
// repository lookup requests by users. So we use a generous 7,200/hr here until we hear
// from someone that these values do not work well for them.
const (
	rateLimitRequestsPerSecond = 2 // 120/min or 7200/hr
	RateLimitMaxBurstRequests  = 500
)

// Client access a Bitbucket Cloud via the REST API 2.0.
type Client struct {
	// HTTP Client used to communicate with the API
	httpClient httpcli.Doer

	// URL is the base URL of Bitbucket Cloud.
	URL *url.URL

	// The username and app password credentials for accessing the server.
	Username, AppPassword string

	// RateLimit is the self-imposed rate limiter (since Bitbucket does not have a concept
	// of rate limiting in HTTP response headers).
	RateLimit *rate.Limiter
}

// NewClient creates a new Bitbucket Cloud API client. If a nil
// httpClient is provided, http.DefaultClient will be used.
func NewClient(httpClient httpcli.Doer) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	httpClient = requestCounter.Doer(httpClient, func(u *url.URL) string {
		// The second component of the Path mostly maps to the type of API
		// request we are making.
		var category string
		if parts := strings.SplitN(u.Path, "/", 4); len(parts) > 2 {
			category = parts[2]
		}
		return category
	})

	return &Client{
		httpClient: httpClient,
		URL:        &url.URL{Scheme: "https", Host: "api.bitbucket.org", Path: "/2.0"},
		RateLimit:  rate.NewLimiter(rateLimitRequestsPerSecond, RateLimitMaxBurstRequests),
	}
}

// Repos returns a list of repositories that are fetched and populated based on user
// authencicated to this client and given pagination criteria. If the argument pageToken.Next
// is not empty, it will be used directly as the URL to make the request. The PageToken it
// returns may also contain the URL to the next page for succeeding requests if any.
func (c *Client) Repos(ctx context.Context, pageToken *PageToken) ([]*Repo, *PageToken, error) {
	var repos []*Repo
	var next *PageToken
	var err error
	if pageToken.HasMore() {
		next, err = c.reqPage(ctx, pageToken.Next, &repos)
	} else {
		next, err = c.page(ctx, fmt.Sprintf("/2.0/repositories/%s", c.Username), nil, pageToken, &repos)
	}
	return repos, next, err
}

// TeamRepos returns a list of repositories that are fetched and populated based on given team
// name and pagination criteria. This includes private repositories, but filtered down to the
// ones that the user authencicated to this client has access to. If the argument pageToken.Next
// is not empty, it will be used directly as the URL to make the request. The PageToken it
// returns may also contain the URL to the next page for succeeding requests if any.
func (c *Client) TeamRepos(ctx context.Context, pageToken *PageToken, teamName string) ([]*Repo, *PageToken, error) {
	var repos []*Repo
	var next *PageToken
	var err error
	if pageToken.HasMore() {
		next, err = c.reqPage(ctx, pageToken.Next, &repos)
	} else {
		next, err = c.page(ctx, fmt.Sprintf("/2.0/teams/%s/repositories", teamName), nil, pageToken, &repos)
	}
	return repos, next, err
}

func (c *Client) page(ctx context.Context, path string, qry url.Values, token *PageToken, results interface{}) (*PageToken, error) {
	if qry == nil {
		qry = make(url.Values)
	}

	for k, vs := range token.Values() {
		qry[k] = append(qry[k], vs...)
	}

	u := url.URL{Path: path, RawQuery: qry.Encode()}
	return c.reqPage(ctx, u.String(), results)
}

// reqPage directly requests resources from given URL assuming all attributes have been
// included in the URL parameter. This is particular useful since the Bitbucket Cloud
// API 2.0 pagination renders the full link of next page in the response.
// See more at https://developer.atlassian.com/bitbucket/api/2/reference/meta/pagination
// However, for the very first request, use method page instead.
func (c *Client) reqPage(ctx context.Context, url string, results interface{}) (*PageToken, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var next PageToken
	err = c.do(ctx, req, &struct {
		*PageToken
		Values interface{} `json:"values"`
	}{
		PageToken: &next,
		Values:    results,
	})

	if err != nil {
		return nil, err
	}

	return &next, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) error {
	req.URL = c.URL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(),
		req.WithContext(ctx),
		nethttp.OperationName("Bitbucket Cloud"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	if err := c.authenticate(req); err != nil {
		return err
	}

	startWait := time.Now()
	if err := c.RateLimit.Wait(ctx); err != nil {
		return err
	}

	if d := time.Since(startWait); d > 200*time.Millisecond {
		log15.Warn("Bitbucket Cloud self-enforced API rate limit: request delayed longer than expected due to rate limit", "delay", d)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return errors.WithStack(&httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		})
	}

	if result != nil {
		return json.Unmarshal(bs, result)
	}

	return nil
}

func (c *Client) authenticate(req *http.Request) error {
	req.SetBasicAuth(c.Username, c.AppPassword)
	return nil
}

type PageToken struct {
	Size    int    `json:"size"`
	Page    int    `json:"page"`
	Pagelen int    `json:"pagelen"`
	Next    string `json:"next"`
}

func (t *PageToken) HasMore() bool {
	if t == nil {
		return true
	}
	return len(t.Next) > 0
}

func (t *PageToken) Values() url.Values {
	v := url.Values{}
	if t == nil {
		return v
	}
	if t.Pagelen != 0 {
		v.Set("pagelen", strconv.Itoa(t.Pagelen))
	}
	return v
}

type Repo struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	UUID        string `json:"uuid"`
	SCM         string `json:"scm"`
	Description string `json:"description"`
	Parent      *Repo  `json:"parent"`
	IsPrivate   bool   `json:"is_private"`
	Links       struct {
		Clone CloneLinks `json:"clone"`
		HTML  Link       `json:"html"`
	} `json:"links"`
}

type CloneLinks []struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type Link struct {
	Href string `json:"href"`
}

// HTTPS returns clone link named "https", it returns an error if not found.
func (cl CloneLinks) HTTPS() (string, error) {
	for _, l := range cl {
		if l.Name == "https" {
			return l.Href, nil
		}
	}
	return "", errors.New("HTTPS clone link not found")
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Bitbucket Cloud API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

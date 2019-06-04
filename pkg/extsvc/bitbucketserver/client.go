package bitbucketserver

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
// for every 1000 repositories on Bitbucket every 1m by default, so for someone with 20,000
// Bitbucket repositories we need 20,000/1000 requests per minute (1200/hr) + overhead for
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

// UserFilters is a list of UserFilter that is ANDed together.
type UserFilters []UserFilter

// EncodeTo encodes the UserFilter to the given url.Values.
func (fs UserFilters) EncodeTo(qry url.Values) {
	var perm int
	for _, f := range fs {
		if f.Filter != "" {
			qry.Set("filter", f.Filter)
		}

		if f.Group != "" {
			qry.Set("group", f.Group)
		}

		if p := f.Permission; p != (PermissionFilter{}) {
			perm++

			q := fmt.Sprintf("permission.%d", perm)
			qry.Set(q, string(p.Root))

			if p.ProjectID != "" {
				qry.Set(q+".projectId", p.ProjectID)
			}

			if p.ProjectKey != "" {
				qry.Set(q+".projectKey", p.ProjectKey)
			}

			if p.RepositoryID != "" {
				qry.Set(q+".repositoryId", p.RepositoryID)
			}

			if p.RepositorySlug != "" {
				qry.Set(q+".repositorySlug", p.RepositorySlug)
			}
		}
	}
}

// UserFilter defines a union type of filters to be used when listing users.
type UserFilter struct {
	// Filter filters the returned users to those whose username,
	// name or email address contain this value.
	// The API doesn't support exact matches.
	Filter string
	// Group filters the returned users to those who are in the give group.
	Group string
	// Permission filters the returned users to those having the given
	// permissions.
	Permission PermissionFilter
}

// A PermissionFilter is a UserFilter used to list users that have specific
// permissions.
type PermissionFilter struct {
	Root           Perm
	ProjectID      string
	ProjectKey     string
	RepositoryID   string
	RepositorySlug string
}

// Users retrieves a page of users, optionally run through provided filters.
func (c *Client) Users(ctx context.Context, pageToken *PageToken, fs ...UserFilter) ([]*User, *PageToken, error) {
	qry := make(url.Values)
	UserFilters(fs).EncodeTo(qry)

	var users []*User
	next, err := c.page(ctx, "rest/api/1.0/users", qry, pageToken, &users)
	return users, next, err
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

func (c *Client) Repos(ctx context.Context, pageToken *PageToken, searchQueries ...string) ([]*Repo, *PageToken, error) {
	qry, err := parseQueryStrings(searchQueries...)
	if err != nil {
		return nil, pageToken, err
	}

	var repos []*Repo
	next, err := c.page(ctx, "rest/api/1.0/repos", qry, pageToken, &repos)
	return repos, next, err
}

func (c *Client) RecentRepos(ctx context.Context, pageToken *PageToken) ([]*Repo, *PageToken, error) {
	var repos []*Repo
	next, err := c.page(ctx, "rest/api/1.0/profile/recent/repos", nil, pageToken, &repos)
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
	req, err := http.NewRequest("GET", u.String(), nil)
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

	// Authenticate request, preferring token.
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else if c.Username != "" || c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(),
		req.WithContext(ctx),
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.WithStack(&httpError{URL: req.URL, StatusCode: resp.StatusCode})
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, result)
}

func parseQueryStrings(qs ...string) (url.Values, error) {
	vals := make(url.Values)
	for _, q := range qs {
		query, err := url.ParseQuery(q)
		if err != nil {
			return nil, err
		}
		for k, vs := range query {
			vals[k] = append(vals[k], vs...)
		}
	}
	return vals, nil
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
	v := t.Values()
	if len(v) == 0 {
		return ""
	}
	return "?" + v.Encode()
}

func (t *PageToken) Values() url.Values {
	v := url.Values{}
	if t == nil {
		return v
	}
	if t.NextPageStart != 0 {
		v.Set("start", strconv.Itoa(t.NextPageStart))
	}
	if t.Limit != 0 {
		v.Set("limit", strconv.Itoa(t.Limit))
	}
	return v
}

// Perm represents a Bitbucket Server permission.
type Perm string

// Permission constants.
const (
	PermSysAdmin      Perm = "SYS_ADMIN"
	PermAdmin         Perm = "ADMIN"
	PermLicensedUser  Perm = "LICENSED_USER"
	PermProjectCreate Perm = "PROJECT_CREATE"

	PermProjectAdmin Perm = "PROJECT_ADMIN"
	PermProjectWrite Perm = "PROJECT_WRITE"
	PermProjectView  Perm = "PROJECT_VIEW"
	PermProjectRead  Perm = "PROJECT_READ"

	PermRepoAdmin Perm = "REPO_ADMIN"
	PermRepoRead  Perm = "REPO_READ"
	PermRepoWrite Perm = "REPO_WRITE"
)

// User account in a Bitbucket Server instance.
type User struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	ID           int    `json:"id"`
	DisplayName  string `json:"displayName"`
	Active       bool   `json:"active"`
	Slug         string `json:"slug"`
	Type         string `json:"type"`
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

// IsNotFound reports whether err is a Bitbucket Server API not found error.
func IsNotFound(err error) bool {
	switch e := errors.Cause(err).(type) {
	case *httpError:
		return e.NotFound()
	}
	return false
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

//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide access to the headers
package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client access a Gerrit via the REST API.
type Client struct {
	// HTTP Client used to communicate with the API
	httpClient httpcli.Doer

	// URL is the base URL of Gerrit.
	URL *url.URL

	// RateLimit is the self-imposed rate limiter (since Gerrit does not have a concept
	// of rate limiting in HTTP response headers).
	rateLimit *ratelimit.InstrumentedLimiter

	// Authenticator used to authenticate HTTP requests.
	auther auth.Authenticator
}

// NewClient returns an authenticated Gerrit API client with
// the provided configuration. If a nil httpClient is provided, httpcli.ExternalDoer
// will be used.
func NewClient(urn string, url *url.URL, creds *AccountCredentials, httpClient httpcli.Doer) (*Client, error) {
	if httpClient == nil {
		httpClient = httpcli.ExternalDoer
	}

	auther := &auth.BasicAuth{
		Username: creds.Username,
		Password: creds.Password,
	}

	return &Client{
		httpClient: httpClient,
		URL:        url,
		rateLimit:  ratelimit.DefaultRegistry.Get(urn),
		auther:     auther,
	}, nil
}

func (c *Client) WithAuthenticator(a auth.Authenticator) *Client {
	return &Client{
		httpClient: c.httpClient,
		URL:        c.URL,
		rateLimit:  c.rateLimit,
		auther:     a,
	}
}

func (c *Client) GetAuthenticatedUserAccount(ctx context.Context) (*Account, error) {
	req, err := http.NewRequest("GET", "a/accounts/self", nil)
	if err != nil {
		return nil, err
	}

	var account Account
	if _, err = c.do(ctx, req, &account); err != nil {
		if httpErr := (&httpError{}); errors.As(err, &httpErr) {
			if httpErr.Unauthorized() {
				return nil, errors.New("Invalid username or password.")
			}
		}

		return nil, err
	}

	return &account, nil
}

func (c *Client) GetGroup(ctx context.Context, groupName string) (Group, error) {

	urlGroup := url.URL{Path: fmt.Sprintf("a/groups/%s", groupName)}

	reqAllAccounts, err := http.NewRequest("GET", urlGroup.String(), nil)

	if err != nil {
		return Group{}, err
	}

	respGetGroup := Group{}
	if _, err = c.do(ctx, reqAllAccounts, &respGetGroup); err != nil {
		return respGetGroup, err
	}
	return respGetGroup, nil
}

// ListProjectsArgs defines options to be set on ListProjects method calls.
type ListProjectsArgs struct {
	Cursor *Pagination
	// If true, only fetches repositories with type CODE
	OnlyCodeProjects bool
}

// ListProjectsResponse defines a response struct returned from ListProjects method calls.
type ListProjectsResponse map[string]*Project

func (c *Client) listCodeProjects(ctx context.Context, cursor *Pagination) (ListProjectsResponse, bool, error) {
	// Unfortunately Gerrit APIs are quite limited and don't support pagination well.
	// e.g. when we request a list of 100 CODE projects, 100 projects are fetched and
	// only then filtered for CODE projects, possibly returning less than 100 projects.
	// This means we cannot rely on the number of projects returned to determine if
	// there are more projects to fetch.
	// Currently, if you want to only get CODE projects and want to know if there is another page
	// to query for, the only way to do that is to query both CODE and ALL projects and compare
	// the number of projects returned.

	query := make(url.Values)
	query.Set("n", strconv.Itoa(cursor.PerPage))
	query.Set("S", strconv.Itoa((cursor.Page-1)*cursor.PerPage))
	query.Set("type", "CODE")

	uProjects := url.URL{Path: "a/projects/", RawQuery: query.Encode()}
	req, err := http.NewRequest("GET", uProjects.String(), nil)
	if err != nil {
		return nil, false, err
	}

	var projects ListProjectsResponse
	if _, err = c.do(ctx, req, &projects); err != nil {
		return nil, false, err
	}

	// If the number of projects returned is zero we cannot assume that there is no next page.
	// We fetch the first project on the next page of ALL projects and check if that page is empty.
	if len(projects) == 0 {
		nextPageProject, _, err := c.listAllProjects(ctx, &Pagination{PerPage: 1, Skip: cursor.Page * cursor.PerPage})
		if err != nil {
			return nil, false, err
		}
		if len(nextPageProject) == 0 {
			return projects, false, nil
		}
	}

	// Otherwise we always assume that there is a next page.
	return projects, true, nil
}

func (c *Client) listAllProjects(ctx context.Context, cursor *Pagination) (ListProjectsResponse, bool, error) {
	query := make(url.Values)
	query.Set("n", strconv.Itoa(cursor.PerPage))
	if cursor.Skip > 0 {
		query.Set("S", strconv.Itoa(cursor.Skip))
	} else {
		query.Set("S", strconv.Itoa((cursor.Page-1)*cursor.PerPage))
	}

	uProjects := url.URL{Path: "a/projects/", RawQuery: query.Encode()}
	req, err := http.NewRequest("GET", uProjects.String(), nil)
	if err != nil {
		return nil, false, err
	}

	var projects ListProjectsResponse
	if _, err = c.do(ctx, req, &projects); err != nil {
		return nil, false, err
	}

	// If the number of returned projects equal the number of requested projects,
	// we assume that there is a next page.
	return projects, len(projects) == cursor.PerPage, nil
}

// ListProjects fetches a list of CODE projects from Gerrit.
func (c *Client) ListProjects(ctx context.Context, opts ListProjectsArgs) (projects ListProjectsResponse, nextPage bool, err error) {

	if opts.Cursor == nil {
		opts.Cursor = &Pagination{PerPage: 100, Page: 1}
	}

	if opts.OnlyCodeProjects {
		return c.listCodeProjects(ctx, opts.Cursor)
	}

	return c.listAllProjects(ctx, opts.Cursor)
}

func (c *Client) do(ctx context.Context, req *http.Request, result any) (*http.Response, error) { //nolint:unparam // http.Response is never used, but it makes sense API wise.
	req.URL = c.URL.ResolveReference(req.URL)

	// Authenticate request with auther
	if c.auther != nil {
		if err := c.auther.Authenticate(req); err != nil {
			return nil, err
		}
	}

	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, &httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		}
	}

	// The first 4 characters of the Gerrit API responses need to be stripped, see: https://gerrit-review.googlesource.com/Documentation/rest-api.html#output .
	if len(bs) < 4 {
		return nil, &httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		}
	}
	return resp, json.Unmarshal(bs[4:], result)
}

type Account struct {
	ID          int32  `json:"_account_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
}

type Group struct {
	ID          string `json:"id"`
	GroupID     int32  `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedOn   string `json:"created_on"`
	Owner       string `json:"owner"`
	OwnerID     string `json:"owner_id"`
}

type Project struct {
	Description string            `json:"description"`
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Parent      string            `json:"parent"`
	State       string            `json:"state"`
	Branches    map[string]string `json:"branches"`
	Labels      map[string]Label  `json:"labels"`
}

type Label struct {
	Values       map[string]string `json:"values"`
	DefaultValue string            `json:"default_value"`
}

type Pagination struct {
	PerPage int
	// Either Skip or Page should be set. If Skip is non-zero, it takes precedence.
	Page int
	Skip int
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Gerrit API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

package generalprotocol

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/metrics"
	"github.com/tomnomnom/linkheader"
)

var requestCounter = metrics.NewRequestCounter("general_protocol_requests_count", "Total number of requests sent via general protocol.")

// Client access a code host that implements general protocol.
type Client struct {
	// HTTP Client used to communicate with the code host.
	httpClient httpcli.Doer

	// URL is the base URL of code host
	URL *url.URL

	// Token is the personal access token for accessing the code host.
	Token string

	// The username and password credentials for accessing the server. Typically these
	// are only used when the server doesn't support personal access tokens. If both
	// Token and Username/Password are specified, Token is used.
	Username, Password string
}

// NewClient creates a new General Protocol client. If a nil
// httpClient is provided, http.DefaultClient will be used.
func NewClient(endpoint *url.URL, httpClient httpcli.Doer) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	// e.g. /srcgraph -> 1 + 1 = 2
	// 		/-/srcgraph -> 2 + 1 = 3
	basePathDepth := strings.Count(endpoint.Path, "/") + 1

	httpClient = requestCounter.Doer(httpClient, func(u *url.URL) string {
		// The first component of the Path maps to the type of request we are making.
		var category string
		// Always split one more than need to only get category level info.
		// e.g. /srcgraph/info -> 2 + 2 = 4
		// 		/-/srcgraph/info -> 3 + 2 = 5
		if parts := strings.SplitN(u.Path, "/", basePathDepth+2); len(parts) > basePathDepth {
			category = parts[basePathDepth]
		}
		return category
	})

	return &Client{
		httpClient: httpClient,
		URL:        endpoint,
	}
}

type Info struct {
	Version    int `json:"version"`
	MaxPageLen int `json:"maxPageLen"`
}

func (c *Client) Info(ctx context.Context) (*Info, error) {
	req, err := http.NewRequest("GET", c.URL.Path+"info", nil)
	if err != nil {
		return nil, err
	}

	info := &Info{}
	err = c.do(ctx, req, nil, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// Repos returns a list of repositories that are fetched and populated based on authenticated
// account and pagination criteria.
// If the argument pageToken.Next is not empty, it will be used directly as the URL to make
// the request. The PageToken it returns may also contain the URL to the next page for
// succeeding requests if any.
func (c *Client) Repos(ctx context.Context, pageToken *PageToken) ([]*Repo, *PageToken, error) {
	return c.repos(ctx, pageToken, "repos")
}

// Repos returns a list of repositories that are fetched and populated based on given account
// name and pagination criteria. If the account requested is an organization, results will be
// filtered down to the ones that the user has access to.
// If the argument pageToken.Next is not empty, it will be used directly as the URL to make
// the request. The PageToken it returns may also contain the URL to the next page for
// succeeding requests if any.
func (c *Client) UserRepos(ctx context.Context, pageToken *PageToken, user string) ([]*Repo, *PageToken, error) {
	return c.repos(ctx, pageToken, fmt.Sprintf("repos/%s", user))
}

func (c *Client) repos(ctx context.Context, pageToken *PageToken, path string) ([]*Repo, *PageToken, error) {
	var repos []*Repo
	var next *PageToken
	var err error
	if pageToken.HasNext() {
		next, err = c.reqPage(ctx, pageToken.Next, &repos)
	} else {
		next, err = c.page(ctx, path, nil, pageToken, &repos)
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

	u := url.URL{Path: c.URL.Path + path, RawQuery: qry.Encode()}
	return c.reqPage(ctx, u.String(), results)
}

// reqPage directly requests resources from given URL assuming all attributes have been
// included in the URL parameter. This is particular useful since the General Protocol
// pagination renders the full link of next page in the response.
// However, for the very first request, use method page instead.
func (c *Client) reqPage(ctx context.Context, url string, results interface{}) (*PageToken, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var next PageToken
	err = c.do(ctx, req, &next, results)
	if err != nil {
		return nil, err
	}

	return &next, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, next *PageToken, result interface{}) error {
	req.URL = c.URL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(),
		req.WithContext(ctx),
		nethttp.OperationName("General Protocol"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	if err := c.authenticate(req); err != nil {
		return err
	}

	// TODO: implement rate limit
	//startWait := time.Now()
	//if err := c.RateLimit.Wait(ctx); err != nil {
	//	return err
	//}
	//
	//if d := time.Since(startWait); d > 200*time.Millisecond {
	//	log15.Warn("Bitbucket Cloud self-enforced API rate limit: request delayed longer than expected due to rate limit", "delay", d)
	//}

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

	if next != nil {
		links := linkheader.Parse(resp.Header.Get("Link"))
		for _, l := range links.FilterByRel("next") {
			u, err := url.Parse(l.URL)
			if err != nil {
				return errors.Errorf("parse link URL %q: %v", l.URL, err)
			}
			next.Page, _ = strconv.Atoi(u.Query().Get("page"))
			if next.Page > 1 {
				next.Page--
			}
			next.PageLen, _ = strconv.Atoi(u.Query().Get("pageLen"))
			next.Next = l.URL
			break
		}
		next.IsLastPage = !next.HasNext() && len(links.FilterByRel("last")) > 0
	}

	if result != nil {
		return json.Unmarshal(bs, result)
	}

	return nil
}

func (c *Client) authenticate(req *http.Request) error {
	// Authenticate request, in order of preference.
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else if c.Username != "" || c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	return nil
}

type PageToken struct {
	Page       int    `json:"page"`
	PageLen    int    `json:"pageLen"`
	Next       string `json:"next"`
	IsLastPage bool   `json:"isLastPage"`
}

func (t *PageToken) HasNext() bool {
	if t == nil {
		return false
	}
	return len(t.Next) > 0
}

func (t *PageToken) Values() url.Values {
	v := url.Values{}
	if t == nil {
		return v
	}
	if t.PageLen != 0 {
		v.Set("pageLen", strconv.Itoa(t.PageLen))
	}
	return v
}

// Repo contains information of a source repository to be consumed.
type Repo struct {
	// The identifier of the repository, could be a string format of numeric ID,
	// or UUID, etc.
	ID string `json:"id"`
	// The name of the repository.
	Name string `json:"name"`
	// The slug is a URL-friendly version of the repository name.
	Slug string `json:"slug"`
	// The full name usually contains both the owner and the repository name,
	// e.g. "alice/repo". If the code host uses flat strucutre, the value of
	// this field would be the same as Name field.
	FullName string `json:"fullName"`
	// The source code management protocol used by the code host for the
	// repository. Currently, repositories that do not use "git" for this field
	// will simply be ignored.
	SCM string `json:"scm"`
	// The description of the repository.
	Description string `json:"description"`
	// The visibility of the repository.
	IsPrivate bool `json:"is_private"`
	// The parent repository object if this is a forked repository.
	Parent *Repo `json:"parent"`
	// The links for the repository, including clone links of different protocols
	// the code host supports.
	Links []Link `json:"links"`
}

// Link contains the type and the actual URL of a link.
type Link struct {
	Type string `json:"name"`
	Href string `json:"href"`
}

// Link returns link with given type, it returns an error if not found.
func (r *Repo) Link(name string) (string, error) {
	for _, l := range r.Links {
		if l.Type == name {
			return l.Href, nil
		}
	}
	return "", errors.Errorf("link type %q not found", name)
}

// HTTPLink returns link with type "http", the actual URL could either be http://
// or https:// based on response from the external service. It returns an error
// if not found.
func (r *Repo) HTTPLink() (string, error) {
	return r.Link("http")
}

// SSHLink returns link with type "ssh", it returns an error if not found.
func (r *Repo) SSHLink() (string, error) {
	return r.Link("ssh")
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("General Protocol HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

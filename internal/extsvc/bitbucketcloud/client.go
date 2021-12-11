package bitbucketcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/segmentio/fasthash/fnv1"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/schema"
)

var requestCounter = metrics.NewRequestMeter("bitbucket_cloud_requests_count", "Total number of requests sent to the Bitbucket Cloud API.")

// These fields define the self-imposed Bitbucket rate limit (since Bitbucket Cloud does
// not have a concept of rate limiting in HTTP response headers).
//
// See https://godoc.org/golang.org/x/time/rate#Limiter for an explanation of these fields.
//
// The limits chosen here are based on the following logic: Bitbucket Cloud restricts
// "List all repositories" requests (which are a good portion of our requests) to 1,000/hr,
// and they restrict "List a user or team's repositories" requests (which are roughly equal
// to our repository lookup requests) to 1,000/hr.
// See `pkg/extsvc/bitbucketserver/client.go` for the calculations behind these limits`
const (
	defaultRateLimitRequestsPerSecond = 2 // 120/min or 7200/hr
	defaultRateLimitMaxBurstRequests  = 500
)

// Client access a Bitbucket Cloud via the REST API 2.0.
type Client struct {
	// HTTP Client used to communicate with the API
	httpClient httpcli.Doer

	// URL is the base URL of Bitbucket Cloud.
	URL *url.URL

	Auth auth.Authenticator

	// RateLimit is the self-imposed rate limiter (since Bitbucket does not have a concept
	// of rate limiting in HTTP response headers).
	RateLimit *rate.Limiter
}

// NewClient creates a new Bitbucket Cloud API client with given apiURL. If a nil httpClient
// is provided, http.DefaultClient will be used. Both Username and AppPassword fields are
// required to be set before calling any APIs.
func NewClient(config *schema.BitbucketCloudConnection, httpClient httpcli.Doer) (*Client, error) {
	if httpClient == nil {
		httpClient = httpcli.ExternalDoer
	}

	apiURL := config.ApiURL
	if apiURL == "" {
		apiURL = "https://api.bitbucket.org"
	}
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	u = extsvc.NormalizeBaseURL(u)

	// TODO: Why wasn't this here before?
	requestsPerHour := rate.Limit(defaultRateLimitRequestsPerSecond)
	if config.RateLimit != nil && config.RateLimit.RequestsPerHour != 0 {
		requestsPerHour = rate.Limit(config.RateLimit.RequestsPerHour)
	}

	// TODO: The categorizer of bitbucket cloud is so different.
	httpClient = requestCounter.Doer(httpClient, func(u *url.URL) string {
		// The second component of the Path mostly maps to the type of API
		// request we are making.
		var category string
		if parts := strings.SplitN(u.Path, "/", 4); len(parts) > 2 {
			category = parts[2]
		}
		return category
	})

	// Normally our registry will return a default infinite limiter when nothing has been
	// synced from config. However, we always want to ensure there is at least some form of rate
	// limiting for Bitbucket.
	defaultLimiter := rate.NewLimiter(requestsPerHour, defaultRateLimitMaxBurstRequests)
	l := ratelimit.DefaultRegistry.GetOrSet(u.String(), defaultLimiter)

	c := &Client{
		httpClient: httpClient,
		URL:        u,
		RateLimit:  l,
	}

	if config.Username != "" && config.AppPassword != "" {
		c.Auth = &auth.BasicAuth{Username: config.Username, Password: config.AppPassword}
	}

	return c, nil
}

// WithAuthenticator returns a new Client that uses the same configuration,
// HTTPClient, and RateLimiter as the current Client, except authenticated user
// with the given authenticator instance.
func (c *Client) WithAuthenticator(a auth.Authenticator) *Client {
	return &Client{
		httpClient: c.httpClient,
		URL:        c.URL,
		RateLimit:  c.RateLimit,
		Auth:       a,
	}
}

// ErrPullRequestNotFound is returned by LoadPullRequest when the pull request has
// been deleted on upstream, or never existed. It will NOT be thrown, if it can't
// be determined whether the pull request exists, because the credential used
// cannot view the repository.
var ErrPullRequestNotFound = errors.New("pull request not found")

// LoadPullRequest loads the given PullRequest returning an error in case of failure.
func (c *Client) LoadPullRequest(ctx context.Context, pr *PullRequest) error {
	if pr.ToRef.Repository.Slug == "" {
		return errors.New("repository slug empty")
	}
	if pr.ToRef.Repository.Project.Key == "" {
		return errors.New("project key empty")
	}

	path := fmt.Sprintf(
		"rest/api/1.0/projects/%s/repos/%s/pull-requests/%d",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)
	_, err := c.send(ctx, "GET", path, nil, nil, pr)
	if err != nil {
		var e *httpError
		if errors.As(err, &e) && e.NoSuchPullRequestException() {
			return ErrPullRequestNotFound
		}

		return err
	}

	return nil
}

// Repos returns a list of repositories that are fetched and populated based on given account
// name and pagination criteria. If the account requested is a team, results will be filtered
// down to the ones that the app password's user has access to.
// If the argument pageToken.Next is not empty, it will be used directly as the URL to make
// the request. The PageToken it returns may also contain the URL to the next page for
// succeeding requests if any.
func (c *Client) Repos(ctx context.Context, pageToken *PageToken, accountName string) ([]*Repo, *PageToken, error) {
	var repos []*Repo
	var next *PageToken
	var err error
	if pageToken.HasMore() {
		next, err = c.reqPage(ctx, pageToken.Next, &repos)
	} else {
		next, err = c.page(ctx, fmt.Sprintf("/2.0/repositories/%s", accountName), nil, pageToken, &repos)
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
	_, err = c.do(ctx, req, &struct {
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

func (c *Client) send(ctx context.Context, method, path string, qry url.Values, payload, result interface{}) (*http.Response, error) {
	if qry == nil {
		qry = make(url.Values)
	}

	var body io.ReadWriter
	if payload != nil {
		body = new(bytes.Buffer)
		if err := json.NewEncoder(body).Encode(payload); err != nil {
			return nil, err
		}
	}

	u := url.URL{Path: path, RawQuery: qry.Encode()}
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req, result)
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) (*http.Response, error) {
	req.URL = c.URL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	req, ht := nethttp.TraceRequest(ot.GetTracer(ctx),
		req.WithContext(ctx),
		nethttp.OperationName("Bitbucket Cloud"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	if err := c.authenticate(req); err != nil {
		return nil, err
	}

	startWait := time.Now()
	if err := c.RateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	if d := time.Since(startWait); d > 200*time.Millisecond {
		log15.Warn("Bitbucket Cloud self-enforced API rate limit: request delayed longer than expected due to rate limit", "delay", d)
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
		return nil, errors.WithStack(&httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		})
	}

	// handle binary response
	if s, ok := result.(*[]byte); ok {
		*s = bs
	} else if result != nil {
		return resp, json.Unmarshal(bs, result)
	}

	return resp, nil
}

func (c *Client) authenticate(req *http.Request) error {
	return c.Auth.Authenticate(req)
}

type PageToken struct {
	Size    int    `json:"size"`
	Page    int    `json:"page"`
	Pagelen int    `json:"pagelen"`
	Next    string `json:"next"`
}

func (t *PageToken) HasMore() bool {
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
	if t.Pagelen != 0 {
		v.Set("pagelen", strconv.Itoa(t.Pagelen))
	}
	return v
}

// User account in a Bitbucket Server instance.
type User struct {
	Name         string `json:"name,omitempty"`
	Password     string `json:"-"`
	EmailAddress string `json:"emailAddress,omitempty"`
	ID           int    `json:"id,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Slug         string `json:"slug,omitempty"`
	Type         string `json:"type,omitempty"`
}

type Project struct {
	UUID   string `json:"uuid"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
	Links  struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type Repo struct {
	UUID        string   `json:"uuid"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	SCM         string   `json:"scm"`
	Description string   `json:"description"`
	Parent      *Repo    `json:"parent"`
	IsPrivate   bool     `json:"is_private"`
	Links       Links    `json:"links"`
	Project     *Project `json:"project"`
}

type Links struct {
	Clone CloneLinks `json:"clone"`
	HTML  Link       `json:"html"`
}

type CloneLinks []struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type Link struct {
	Href string `json:"href"`
}

type Ref struct {
	ID         string `json:"id"`
	Repository struct {
		ID      int    `json:"id"`
		Slug    string `json:"slug"`
		Project struct {
			Key string `json:"key"`
		} `json:"project"`
	} `json:"repository"`
}

type PullRequest struct {
	ID           int               `json:"id"` // OK
	Version      int               `json:"version"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	State        string            `json:"state"`
	Open         bool              `json:"open"`
	Closed       bool              `json:"closed"`
	CreatedDate  int               `json:"createdDate"`
	UpdatedDate  int               `json:"updatedDate"`
	FromRef      Ref               `json:"fromRef"`
	ToRef        Ref               `json:"toRef"`
	Locked       bool              `json:"locked"`
	Author       PullRequestAuthor `json:"author"`
	Reviewers    []Reviewer        `json:"reviewers"`
	Participants []Participant     `json:"participants"`
	Links        struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`

	Activities   []*Activity     `json:"activities,omitempty"`
	Commits      []*Commit       `json:"commits,omitempty"`
	CommitStatus []*CommitStatus `json:"commit_status,omitempty"`

	// Deprecated, use CommitStatus instead. BuildStatus was not tied to individual commits
	BuildStatuses []*BuildStatus `json:"buildstatuses,omitempty"`
}

// PullRequestAuthor is the author of a pull request.
type PullRequestAuthor struct {
	User     *User  `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"approved"`
	Status   string `json:"status"`
}

// Reviewer is a user that left feedback on a pull request.
type Reviewer struct {
	User               *User  `json:"user"`
	LastReviewedCommit string `json:"lastReviewedCommit"`
	Role               string `json:"role"`
	Approved           bool   `json:"approved"`
	Status             string `json:"status"`
}

// Participant is a user that was involved in a pull request.
type Participant struct {
	User     *User  `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"approved"`
	Status   string `json:"status"`
}

// Activity is a union type of all supported pull request activity items.
type Activity struct {
	ID          int            `json:"id"`
	CreatedDate int            `json:"createdDate"`
	User        User           `json:"user"`
	Action      ActivityAction `json:"action"`

	// Comment activity fields.
	CommentAction string         `json:"commentAction,omitempty"`
	Comment       *Comment       `json:"comment,omitempty"`
	CommentAnchor *CommentAnchor `json:"commentAnchor,omitempty"`

	// Reviewers change fields.
	AddedReviewers   []User `json:"addedReviewers,omitempty"`
	RemovedReviewers []User `json:"removedReviewers,omitempty"`

	// Merged event fields.
	Commit *Commit `json:"commit,omitempty"`
}

// Key is a unique key identifying this activity in the context of its pull request.
func (a *Activity) Key() string { return strconv.Itoa(a.ID) }

// BuildStatus represents the build status of a commit
type BuildStatus struct {
	State       string `json:"state,omitempty"`
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	Url         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	DateAdded   int64  `json:"dateAdded,omitempty"`
}

// Commit status is the build status for a specific commit
type CommitStatus struct {
	Commit string      `json:"commit,omitempty"`
	Status BuildStatus `json:"status,omitempty"`
}

func (s *CommitStatus) Key() string {
	key := fmt.Sprintf("%s:%s:%s:%s", s.Commit, s.Status.Key, s.Status.Name, s.Status.Url)
	return strconv.FormatInt(int64(fnv1.HashString64(key)), 16)
}

// ActivityAction defines the action taken in an Activity.
type ActivityAction string

// Known ActivityActions
const (
	ApprovedActivityAction   ActivityAction = "APPROVED"
	UnapprovedActivityAction ActivityAction = "UNAPPROVED"
	DeclinedActivityAction   ActivityAction = "DECLINED"
	ReviewedActivityAction   ActivityAction = "REVIEWED"
	OpenedActivityAction     ActivityAction = "OPENED"
	ReopenedActivityAction   ActivityAction = "REOPENED"
	RescopedActivityAction   ActivityAction = "RESCOPED"
	UpdatedActivityAction    ActivityAction = "UPDATED"
	CommentedActivityAction  ActivityAction = "COMMENTED"
	MergedActivityAction     ActivityAction = "MERGED"
)

// A Comment in a PullRequest.
type Comment struct {
	ID                  int                 `json:"id"`
	Version             int                 `json:"version"`
	Text                string              `json:"text"`
	Author              User                `json:"author"`
	CreatedDate         int                 `json:"createdDate"`
	UpdatedDate         int                 `json:"updatedDate"`
	Comments            []Comment           `json:"comments"` // Replies to the comment
	Tasks               []Task              `json:"tasks"`
	PermittedOperations PermittedOperations `json:"permittedOperations"`
}

// A CommentAnchor captures the location of a code comment in a PullRequest.
type CommentAnchor struct {
	FromHash string `json:"fromHash"`
	ToHash   string `json:"toHash"`
	Line     int    `json:"line"`
	LineType string `json:"lineType"`
	FileType string `json:"fileType"`
	Path     string `json:"path"`
	DiffType string `json:"diffType"`
	Orphaned bool   `json:"orphaned"`
}

// A Task in a PullRequest.
type Task struct {
	ID                  int                 `json:"id"`
	Author              User                `json:"author"`
	Text                string              `json:"text"`
	State               string              `json:"state"`
	CreatedDate         int                 `json:"createdDate"`
	PermittedOperations PermittedOperations `json:"permittedOperations"`
}

// PermittedOperations of a Comment or Task.
type PermittedOperations struct {
	Editable       bool `json:"editable,omitempty"`
	Deletable      bool `json:"deletable,omitempty"`
	Transitionable bool `json:"transitionable,omitempty"`
}

// A Commit in a Repository.
type Commit struct {
	ID                 string   `json:"id,omitempty"`
	DisplayID          string   `json:"displayId,omitempty"`
	Author             *User    `json:"user,omitempty"`
	AuthorTimestamp    int64    `json:"authorTimestamp,omitempty"`
	Committer          *User    `json:"committer,omitempty"`
	CommitterTimestamp int64    `json:"committerTimestamp,omitempty"`
	Message            string   `json:"message,omitempty"`
	Parents            []Commit `json:"parents,omitempty"`
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

const (
	bitbucketDuplicatePRException       = "com.atlassian.bitbucket.pull.DuplicatePullRequestException"
	bitbucketNoSuchLabelException       = "com.atlassian.bitbucket.label.NoSuchLabelException"
	bitbucketNoSuchPullRequestException = "com.atlassian.bitbucket.pull.NoSuchPullRequestException"
)

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

func (e *httpError) DuplicatePullRequest() bool {
	return strings.Contains(string(e.Body), bitbucketDuplicatePRException)
}

func (e *httpError) NoSuchPullRequestException() bool {
	return strings.Contains(string(e.Body), bitbucketNoSuchPullRequestException)
}

package buildkite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/google/go-querystring/query"
)

const (
	defaultBaseURL = "https://api.buildkite.com/"
	userAgent      = "go-buildkite/" + Version
)

var (
	httpDebug = false
)

// Client - A Client manages communication with the buildkite API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests.  Defaults to the public buildkite API. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the buildkite API.
	UserAgent string

	// Services used for talking to different parts of the buildkite API.
	AccessTokens      *AccessTokensService
	Agents            *AgentsService
	Annotations       *AnnotationsService
	Artifacts         *ArtifactsService
	Builds            *BuildsService
	Clusters          *ClustersService
	ClusterQueues     *ClusterQueuesService
	ClusterTokens     *ClusterTokensService
	FlakyTests        *FlakyTestsService
	Jobs              *JobsService
	Organizations     *OrganizationsService
	Pipelines         *PipelinesService
	PipelineTemplates *PipelineTemplatesService
	User              *UserService
	Teams             *TeamsService
	Tests             *TestsService
	TestRuns          *TestRunsService
	TestSuites        *TestSuitesService
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	// For paginated result sets, page of results to retrieve.
	Page int `url:"page,omitempty"`

	// For paginated result sets, the number of results to include per page.
	PerPage int `url:"per_page,omitempty"`
}

// NewClient returns a new buildkite API client. As API calls require authentication
// you MUST supply a client which provides the required API key.
func NewClient(httpClient *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		UserAgent: userAgent,
	}
	c.AccessTokens = &AccessTokensService{c}
	c.Agents = &AgentsService{c}
	c.Annotations = &AnnotationsService{c}
	c.Artifacts = &ArtifactsService{c}
	c.Builds = &BuildsService{c}
	c.Clusters = &ClustersService{c}
	c.ClusterQueues = &ClusterQueuesService{c}
	c.ClusterTokens = &ClusterTokensService{c}
	c.FlakyTests = &FlakyTestsService{c}
	c.Jobs = &JobsService{c}
	c.Organizations = &OrganizationsService{c}
	c.Pipelines = &PipelinesService{c}
	c.PipelineTemplates = &PipelineTemplatesService{c}
	c.User = &UserService{c}
	c.Teams = &TeamsService{c}
	c.Tests = &TestsService{c}
	c.TestRuns = &TestRunsService{c}
	c.TestSuites = &TestSuitesService{c}

	if c.client != nil {
		if tokenAuth, ok := c.client.Transport.(*TokenAuthTransport); ok {
			tokenAuth.APIHost = baseURL.Host
		}

		if basicAuth, ok := c.client.Transport.(*BasicAuthTransport); ok {
			basicAuth.APIHost = baseURL.Host
		}
	}
	return c
}

// SetHttpDebug this enables global http request/response dumping for this API
func SetHttpDebug(flag bool) {
	httpDebug = flag
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.  If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}

	return req, nil
}

// Response is a buildkite API response.  This wraps the standard http.Response
// returned from buildkite and provides convenient access to things like
// pagination links.
type Response struct {
	*http.Response

	// These fields provide the page values for paginating through a set of
	// results.  Any or all of these may be set to the zero value for
	// responses that are not part of a paginated set, or for which there
	// are no additional pages.

	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int
}

// newResponse creats a new Response for the provided http.Response.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	response.populatePageValues()
	return response
}

// populatePageValues parses the HTTP Link response headers and populates the
// various pagination link values in the Reponse.
func (r *Response) populatePageValues() {
	if links, ok := r.Response.Header["Link"]; ok && len(links) > 0 {
		for _, link := range strings.Split(links[0], ",") {
			segments := strings.Split(strings.TrimSpace(link), ";")

			// link must at least have href and rel
			if len(segments) < 2 {
				continue
			}

			// ensure href is properly formatted
			if !strings.HasPrefix(segments[0], "<") || !strings.HasSuffix(segments[0], ">") {
				continue
			}

			// try to pull out page parameter
			url, err := url.Parse(segments[0][1 : len(segments[0])-1])
			if err != nil {
				continue
			}
			page := url.Query().Get("page")
			if page == "" {
				continue
			}

			for _, segment := range segments[1:] {
				switch strings.TrimSpace(segment) {
				case `rel="next"`:
					r.NextPage, _ = strconv.Atoi(page)
				case `rel="prev"`:
					r.PrevPage, _ = strconv.Atoi(page)
				case `rel="first"`:
					r.FirstPage, _ = strconv.Atoi(page)
				case `rel="last"`:
					r.LastPage, _ = strconv.Atoi(page)
				}

			}
		}
	}
}

// Do sends an API request and returns the API response.  The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.  If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
	respCh := make(chan *http.Response, 1)

	op := func() error {
		if httpDebug {
			if dump, err := httputil.DumpRequest(req, true); err == nil {
				fmt.Printf("DEBUG request uri=%s\n%s\n", req.URL, dump)
			}
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return backoff.Permanent(err)
		}

		if httpDebug {
			if dump, err := httputil.DumpResponse(resp, true); err == nil {
				fmt.Printf("DEBUG response uri=%s\n%s\n", req.URL, dump)
			}
		}

		// Check for rate limiting response on idempotent requests
		if req.Method == http.MethodGet && resp.StatusCode == http.StatusTooManyRequests {
			errMsg := resp.Header.Get("Rate-Limit-Warning")
			if errMsg == "" {
				errMsg = "Too many requests, retry"
			}
			return errors.New(errMsg)
		}

		respCh <- resp
		return nil
	}

	notify := func(err error, delay time.Duration) {
		if httpDebug {
			fmt.Printf("DEBUG error %v, retry in %v\n", err, delay)
		}
	}

	if err := backoff.RetryNotify(op, backoff.NewExponentialBackOff(), notify); err != nil {
		return nil, err
	}

	resp := <-respCh

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	response := newResponse(resp)

	if err := checkResponse(resp); err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return response, err
	}

	var err error

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return response, err
}

// ErrorResponse provides a message.
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message" yaml:"message"` // error message
	RawBody  []byte         `json:"-" yaml:"-"`             // Raw Response Body
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	data, err := ioutil.ReadAll(r.Body)
	errorResponse := &ErrorResponse{Response: r, RawBody: data}
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

// addOptions adds the parameters in opt as URL query parameters to s.  opt
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it, but unlike Int
// its argument value is an int.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

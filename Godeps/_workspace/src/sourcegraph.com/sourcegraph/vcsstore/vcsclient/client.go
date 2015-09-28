package vcsclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/git"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "vcsstore-client/" + libraryVersion
)

// A Client communicates with the vcsstore API.
type Client struct {
	// Base URL for API requests, which should have a trailing slash.
	BaseURL *url.URL

	// User agent used for HTTP requests to the vcsstore API.
	UserAgent string

	// HTTP client used to communicate with the vcsstore API.
	httpClient *http.Client

	// HTTP client that is identical to httpClient except it does not follow
	// redirects.
	ignoreRedirectsHTTPClient *http.Client
}

var _ VCSStore = (*Client)(nil)

// New returns a new vcsstore API client that communicates with an HTTP server
// at the base URL. If httpClient is nil, http.DefaultClient is used.
func New(base *url.URL, httpClient *http.Client) *Client {
	if httpClient == nil {
		cloned := *http.DefaultClient
		httpClient = &cloned
	}

	ignoreRedirectsHTTPClient := *httpClient
	ignoreRedirectsHTTPClient.CheckRedirect = func(r *http.Request, via []*http.Request) error { return errIgnoredRedirect }

	c := &Client{
		BaseURL:                   base,
		UserAgent:                 userAgent,
		httpClient:                httpClient,
		ignoreRedirectsHTTPClient: &ignoreRedirectsHTTPClient,
	}

	return c
}

func (c *Client) Repository(repoPath string) (vcs.Repository, error) {
	return &repository{
		client:   c,
		repoPath: repoPath,
	}, nil
}

func (c *Client) GitTransport(repoPath string) (git.GitTransport, error) {
	return &gitTransport{client: c, repoPath: repoPath}, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client. Relative
// URLs should always be specified without a preceding slash. If specified, the
// value pointed to by body is JSON encoded and included as the request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if c.BaseURL != nil {
		u = c.BaseURL.ResolveReference(u)
	} else {
		u.Path = "/" + u.Path
	}

	var buf bytes.Buffer
	var hasJSONBody bool
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		if err != nil {
			return nil, err
		}
		hasJSONBody = true
	}

	req, err := http.NewRequest(method, u.String(), &buf)
	if err != nil {
		return nil, err
	}

	if hasJSONBody {
		req.Header.Set("content-type", "application/json; charset=utf-8")
	}

	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// Do sends an API request and returns the API response.  The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp, false)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		if bp, ok := v.(*[]byte); ok {
			*bp, err = ioutil.ReadAll(resp.Body)
		} else if buf, ok := v.(*bytes.Buffer); ok {
			_, err = io.Copy(buf, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s %s: %s", req.Method, req.URL.RequestURI(), err)
	}
	return resp, nil
}

// doIgnoringRedirects sends an API request and returns the HTTP response. If
// it encounters an HTTP redirect, it does not follow it.
func (c *Client) doIgnoringRedirects(req *http.Request) (*http.Response, error) {
	resp, err := c.ignoreRedirectsHTTPClient.Do(req)
	if err != nil && !isIgnoredRedirectErr(err) {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, CheckResponse(resp, true)
}

var errIgnoredRedirect = errors.New("not following redirect")

func isIgnoredRedirectErr(err error) bool {
	if err, ok := err.(*url.Error); ok && err.Err == errIgnoredRedirect {
		return true
	}
	return false
}

type RepoKey struct {
	// URI identifies the repository, and it has the form "host/user/repo".
	// Use RepoKey.RepoKeyURI to parse URI as a RepoKeyURI.
	URI string
	VCS string
}

type RepositoryOpener interface {
	Repository(repoPath string) (vcs.Repository, error)
}

type MockRepositoryOpener struct{ Return vcs.Repository }

var _ RepositoryOpener = MockRepositoryOpener{}

func (m MockRepositoryOpener) Repository(repoPath string) (vcs.Repository, error) {
	return m.Return, nil
}

// GetFile gets a file from the repository's tree at a specific commit. If the
// path does not refer to a file, a non-nil error is returned.
func GetFile(o RepositoryOpener, repoPath string, at vcs.CommitID, path string) ([]byte, os.FileInfo, error) {
	r, err := o.Repository(repoPath)
	if err != nil {
		return nil, nil, err
	}

	fs, err := r.FileSystem(at)
	if err != nil {
		return nil, nil, err
	}

	fi, err := fs.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !fi.Mode().IsRegular() {
		return nil, fi, errors.New("tree entry is not a file")
	}

	f, err := fs.Open(path)
	if err != nil {
		return nil, fi, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fi, err
	}

	return data, fi, nil
}

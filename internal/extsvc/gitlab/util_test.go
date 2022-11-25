package gitlab

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

type mockHTTPResponseBody struct {
	count        int
	header       http.Header
	responseBody string
	statusCode   int
}

func (s *mockHTTPResponseBody) Do(req *http.Request) (*http.Response, error) {
	s.count++
	statusCode := http.StatusOK
	if s.statusCode != 0 {
		statusCode = s.statusCode
	}
	return &http.Response{
		Request:    req,
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(s.responseBody)),
		Header:     s.header,
	}, nil
}

type mockHTTPEmptyResponse struct {
	statusCode int
}

func (s mockHTTPEmptyResponse) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Request:    req,
		StatusCode: s.statusCode,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}, nil
}

type mockHTTPPaginatedResponse struct {
	pages []*mockHTTPResponseBody
	empty mockHTTPEmptyResponse
}

func (s *mockHTTPPaginatedResponse) Do(req *http.Request) (*http.Response, error) {
	if len(s.pages) == 0 {
		// If s.empty is the zero value, then we should really return 200.
		if s.empty.statusCode == 0 {
			s.empty.statusCode = http.StatusOK
		}
		return s.empty.Do(req)
	}

	page := s.pages[0]
	s.pages = s.pages[1:]
	return page.Do(req)
}

func newTestClient(t *testing.T) *Client {
	rcache.SetupForTest(t)
	return &Client{
		baseURL:          &url.URL{Scheme: "https", Host: "example.com", Path: "/"},
		httpClient:       &http.Client{},
		rateLimitMonitor: &ratelimit.Monitor{},
		projCache:        rcache.NewWithTTL("__test__gl_proj", 1000),
	}
}

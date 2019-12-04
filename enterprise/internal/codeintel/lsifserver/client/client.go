package client

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver"
	"github.com/tomnomnom/linkheader"
)

var DefaultClient = &Client{
	URL: lsifserver.ServerURLFromEnv,
	HTTPClient: &http.Client{
		// nethttp.Transport will propagate opentracing spans
		Transport: &nethttp.Transport{},
	},
}

type Client struct {
	URL        string
	HTTPClient *http.Client
}

// BuildAndTraceRequest builds a URL and performs a request. This is a convenience wrapper
// around BuildURL and TraceRequest.
func (c *Client) BuildAndTraceRequest(ctx context.Context, method, path string, query url.Values, body io.ReadCloser) (*http.Response, error) {
	url, err := buildURL(c.URL, path, query)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, method, url, body)
}

// TraceRequestAndUnmarshalPayload builds a URL, performs a request, and populates
// the given payload with the response body. This is a convenience wrapper around
// BuildURL, TraceRequest, and UnmarshalPayload.
func (c *Client) TraceRequestAndUnmarshalPayload(ctx context.Context, method, path string, query url.Values, body io.ReadCloser, payload interface{}) error {
	resp, err := c.BuildAndTraceRequest(ctx, method, path, query, body)
	if err != nil {
		return err
	}

	return UnmarshalPayload(resp, &payload)
}

// buildURL constructs a URL to the backend LSIF server with the given path
// and query values. If path is relative (indicated by a leading slash), then
// the configured LSIF server url is prepended. Otherwise, it is treated as
// an absolute URL. The given query values will override any query string that
// is present in the given path.
//
// This method can be used to construct a LSIF request URL either from a root
// relative path on the first request of a paginated endpoint or from the URL
// provided by the Link header in a previous response.
func buildURL(baseURL, path string, query url.Values) (string, error) {
	build := url.Parse
	if len(path) > 0 && path[0] == '/' {
		build = func(path string) (*url.URL, error) {
			u, err := url.Parse(baseURL)
			if err != nil {
				return nil, err
			}

			u.Path = path
			return u, nil
		}
	}

	u, err := build(path)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, values := range query {
		q.Set(key, values[0])
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// do performs a GET request to the given URL with the given context. The
// response is expected to have a 200-level status code. If an error is returned, the
// HTTP response body has been closed.
func (c *Client) do(ctx context.Context, method, url string, body io.ReadCloser) (resp *http.Response, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "lsifserver.client.do")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(
		span.Tracer(),
		req,
		nethttp.OperationName("LSIF client"),
		nethttp.ClientTrace(false),
	)
	defer ht.Finish()

	// Do not use ctxhttp.Do here as it will re-wrap the request
	// with a context and this will causes the ot-headers not to
	// propagate correctly.
	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, errors.Wrap(err, "lsif request failed")
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = errors.WithStack(&lsifError{StatusCode: resp.StatusCode, Message: string(content)})
	return
}

// UnmarshalPayload reads (and closes) the given response body and populates
// the given payload with the JSON response.
func UnmarshalPayload(resp *http.Response, payload interface{}) error {
	defer resp.Body.Close()

	if payload == nil {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(body), &payload)
}

// ExtractNextURL retrieves the URL with rel="next" in the given response's Link
// header. If the link header is empty or has no rel="next", this method returns an
// empty string.
func ExtractNextURL(resp *http.Response) string {
	for _, link := range linkheader.Parse(resp.Header.Get("Link")) {
		if link.Rel == "next" {
			return link.URL
		}
	}

	return ""
}

type lsifError struct {
	StatusCode int
	Message    string
}

func (e *lsifError) Error() string {
	return e.Message
}

func IsNotFound(err error) bool {
	if e, ok := errors.Cause(err).(*lsifError); ok {
		return e.StatusCode == http.StatusNotFound
	}

	return false
}

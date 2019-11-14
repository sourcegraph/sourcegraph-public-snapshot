package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/codeintel/lsifserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/tomnomnom/linkheader"
	"golang.org/x/net/context/ctxhttp"
)

var httpClient = &http.Client{
	// nethttp.Transport will propagate opentracing spans
	Transport: &nethttp.Transport{},
}

// BuildAndTraceRequest builds a URL and performs a request. This is a convenience wrapper
// around BuildURL and TraceRequest.
func BuildAndTraceRequest(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	url, err := buildURL(path, query)
	if err != nil {
		return nil, err
	}

	return traceRequest(ctx, url)
}

// TraceRequestAndUnmarshalPayload builds a URL, performs a request, and populates
// the given payload with the response body. This is a convenience wrapper around
// BuildURL, TraceRequest, and UnmarshalPayload.
func TraceRequestAndUnmarshalPayload(ctx context.Context, path string, query url.Values, payload interface{}) error {
	resp, err := BuildAndTraceRequest(ctx, path, query)
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
func buildURL(path string, query url.Values) (string, error) {
	build := url.Parse
	if len(path) > 0 && path[0] == '/' {
		build = func(path string) (*url.URL, error) {
			u, err := url.Parse(lsifserver.ServerURLFromEnv)
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

// traceRequest performs a GET request to the given URL with the given context. The
// response is expected to have a 200-level status code. If an error is returned, the
// HTTP response body has been closed.
func traceRequest(ctx context.Context, url string) (resp *http.Response, err error) {
	tr, ctx := trace.New(ctx, "lsifRequest", fmt.Sprintf("url: %s", url))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req = req.WithContext(ctx)

	req, ht := nethttp.TraceRequest(
		opentracing.GlobalTracer(),
		req,
		nethttp.OperationName("LSIF client"),
		nethttp.ClientTrace(false),
	)
	defer ht.Finish()

	resp, err = ctxhttp.Do(ctx, httpClient, req)
	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		err = errors.Wrap(err, "lsif request failed")
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = errors.WithStack(&lsifError{StatusCode: resp.StatusCode, Message: string(body)})
	return
}

// UnmarshalPayload reads (and closes) the given response body and populates
// the given payload with the JSON response.
func UnmarshalPayload(resp *http.Response, payload interface{}) error {
	defer resp.Body.Close()

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

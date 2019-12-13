package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/tomnomnom/linkheader"
)

type lsifRequest struct {
	method string
	path   string
	cursor *string
	query  queryValues
	body   io.ReadCloser
}

type lsifResponseMeta struct {
	statusCode int
	nextURL    string
}

// do will make a request to LSIF server. This method will return an error if the request
// cannot be made or the status code is 400 or 500-level. If a non-nil payload is given,
// the request body will be unmarshalled into it.
func (c *Client) do(ctx context.Context, lsifRequest *lsifRequest, payload interface{}) (*lsifResponseMeta, error) {
	method := lsifRequest.method
	if method == "" {
		method = "GET"
	}

	url, err := buildURL(c.URL, lsifRequest.path, lsifRequest.cursor, lsifRequest.query)
	if err != nil {
		return nil, err
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "lsifserver.client.do")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	req, err := http.NewRequest(method, url, lsifRequest.body)
	if err != nil {
		return nil, err
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
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, errors.Wrap(err, "lsif request failed")
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, errors.WithStack(&lsifError{
			StatusCode: resp.StatusCode,
			Message:    string(content),
		})
	}

	if payload != nil {
		if err := json.Unmarshal(content, &payload); err != nil {
			return nil, err
		}
	}

	return &lsifResponseMeta{
		statusCode: resp.StatusCode,
		nextURL:    extractNextURL(resp),
	}, nil
}

// buildURL constructs a URL to the backend LSIF server with the given path
// and query values. If the provided cursor is non-nil, that will be used
// instead of the given path. If path is relative (indicated by a leading slash),
// then the configured LSIF server url is prepended. Otherwise, it is treated as
// an absolute URL. The given query values will override any query string that
// is present in the given path.
//
// This method can be used to construct a LSIF request URL either from a root
// relative path on the first request of a paginated endpoint or from the URL
// provided by the Link header in a previous response.
func buildURL(baseURL, path string, cursor *string, query queryValues) (string, error) {
	if cursor != nil {
		path = *cursor
	}

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

// extractNextURL retrieves the URL with rel="next" in the given response's Link
// header. If the link header is empty or has no rel="next", this method returns an
// empty string.
func extractNextURL(resp *http.Response) string {
	for _, link := range linkheader.Parse(resp.Header.Get("Link")) {
		if link.Rel == "next" {
			return link.URL
		}
	}

	return ""
}

type queryValues url.Values

func (qv queryValues) Set(name string, value string) {
	qv[name] = []string{value}
}

func (qv queryValues) SetInt(name string, value int64) {
	qv.Set(name, strconv.FormatInt(int64(value), 10))
}

func (qv queryValues) SetOptionalString(name string, value *string) {
	if value != nil {
		qv.Set(name, *value)
	}
}

func (qv queryValues) SetOptionalInt32(name string, value *int32) {
	if value != nil {
		qv.SetInt(name, int64(*value))
	}
}

func (qv queryValues) SetOptionalBool(name string, value *bool) {
	if value != nil {
		qv.Set(name, fmt.Sprintf("%v", *value))
	}
}

package gosyntect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Query represents a code highlighting query to the syntect_server.
type Query struct {
	// Extension is the file extension of the code.
	//
	// See https://github.com/sourcegraph/syntect_server#supported-file-extensions
	Extension string `json:"extension"`

	// Theme is the color theme to use for highlighting.
	//
	// See https://github.com/sourcegraph/syntect_server#embedded-themes
	Theme string `json:"theme"`

	// Code is the literal code to highlight.
	Code string `json:"code"`
}

// Response represents a response to a code highlighting query.
type Response struct {
	// Data is the actual highlighted HTML version of Query.Code.
	Data string
}

// Error is an error returned from the syntect_server.
type Error string

func (e Error) Error() string {
	return string(e)
}

type response struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

// Client represents a client connection to a syntect_server.
type Client struct {
	syntectServer string
}

// Highlight performs a query to highlight some code.
func (c *Client) Highlight(ctx context.Context, q *Query) (*Response, error) {
	// Build the request.
	jsonQuery, err := json.Marshal(q)
	if err != nil {
		return nil, errors.Wrap(err, "encoding query")
	}
	req, err := http.NewRequest("POST", c.url("/"), bytes.NewReader(jsonQuery))
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}
	req.Header.Set("Content-Type", "application/json")

	// Add tracing to the request.
	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Highlight"),
		nethttp.ClientTrace(false))
	defer ht.Finish()
	client := &http.Client{Transport: &nethttp.Transport{}}

	// Perform the request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("making request to %s", c.url("/")))
	}
	defer resp.Body.Close()

	// Can only call ht.Span() after the request has been exected, so add our span tags in now.
	ht.Span().SetTag("Extension", q.Extension)
	ht.Span().SetTag("Theme", q.Theme)

	// Decode the response.
	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("decoding JSON response from %s", c.url("/")))
	}
	if r.Error != "" {
		return nil, errors.Wrap(Error(r.Error), c.syntectServer)
	}
	return &Response{
		Data: r.Data,
	}, nil
}

func (c *Client) url(path string) string {
	return c.syntectServer + path
}

// New returns a client connection to a syntect_server.
func New(syntectServer string) *Client {
	return &Client{
		syntectServer: strings.TrimSuffix(syntectServer, "/"),
	}
}

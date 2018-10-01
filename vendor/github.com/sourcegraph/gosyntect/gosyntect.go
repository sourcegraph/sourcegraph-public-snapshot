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
	// Extension is deprecated: use Filepath instead.
	Extension string `json:"extension"`

	// Filepath is the file path of the code. It can be the full file path, or
	// just the name and extension.
	//
	// See: https://github.com/sourcegraph/syntect_server#supported-file-extensions
	Filepath string `json:"filepath"`

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

	// Plaintext indicates whether or not a syntax could not be found for the
	// file and instead it was rendered as plain text.
	Plaintext bool
}

var (
	// ErrInvalidTheme is returned when the Query.Theme is not a valid theme.
	ErrInvalidTheme = errors.New("invalid theme")
)

type response struct {
	// Successful response fields.
	Data      string `json:"data"`
	Plaintext bool   `json:"plaintext"`

	// Error response fields.
	Error string `json:"error"`
	Code  string `json:"code"`
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
	ht.Span().SetTag("Filepath", q.Filepath)
	ht.Span().SetTag("Theme", q.Theme)

	// Decode the response.
	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("decoding JSON response from %s", c.url("/")))
	}
	if r.Error != "" {
		var err error
		switch r.Code {
		case "invalid_theme":
			err = ErrInvalidTheme
		case "resource_not_found":
			// resource_not_found is returned in the event of a 404, indicating a bug
			// in gosyntect.
			err = errors.New("gosyntect internal error: resource_not_found")
		default:
			err = fmt.Errorf("unknown error=%q code=%q", r.Error, r.Code)
		}
		return nil, errors.Wrap(err, c.syntectServer)
	}
	return &Response{
		Data:      r.Data,
		Plaintext: r.Plaintext,
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

// Package zoekt provides a client to github.com/google/zoekt
package zoekt

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

// The following structs are copied from github.com/google/zoekt/rest/api.go

// SearchRequest is the entry point for the /api/search POST endpoint.
type SearchRequest struct {
	Query string

	// A list of OR'd restrictions.
	Restrict []SearchRequestRestriction
}

// A REST search query must provide a restriction.
type SearchRequestRestriction struct {
	Repo     string
	Branches []string
}

// SearchResponse is the return type for /api/search endpoint
type SearchResponse struct {
	Files []*SearchResponseFile
	Error *string
}

// SearchResponseFile holds the matches within a single file.
type SearchResponseFile struct {
	Repo     string
	Branches []string
	FileName string
	Lines    []*SearchResponseLine
}

// SearchResponseLine holds the matches within a single line.
type SearchResponseLine struct {
	LineNumber int
	Line       string
	Matches    []*SearchResponseMatch
}

// SearchResponseMatch is the matching segment of the line.
type SearchResponseMatch struct {
	// Start of match, in (unicode) characters.
	Start int

	// End of match, in (unicode) characters.
	End int
}

// ListRequest is the entry point for the /api/list POST endpoint.
type ListRequest struct {
	// A list of OR'd restrictions.
	Restrict []ListRequestRestriction
}

type ListRequestRestriction struct {
	Repo string
}

// ListResponse is the return type for /api/search endpoint
type ListResponse struct {
	Repos []*ListResponseRepo
	Error *string
}

// ListResponseRepo holds repository metadata.
type ListResponseRepo struct {
	// Name is the repository name.
	Name string

	// Branches is the branches indexed in this repo.
	Branches []ListResponseBranch
}

// ListResponseBranch describes an indexed branch, which is a name combined
// with a version.
type ListResponseBranch struct {
	Name    string
	Version string
}

// Client is a zoekt client to the zoekt rest API.
type Client struct {
	// Host is the hostname of the zoekt instance. It can include a port. For
	// example "localhost:6070".
	Host string
}

// Search sends a search request. Read the documentation for SearchRequest and
// SearchResponse.
func (c *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	var resp SearchResponse
	err := c.do(ctx, "search", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search zoekt")
	}
	return &resp, nil
}

// List sends a list request.
func (c *Client) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	var resp ListResponse
	err := c.do(ctx, "list", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list zoekt")
	}
	return &resp, nil
}

func (c *Client) do(ctx context.Context, method string, reqBody, respBody interface{}) error {
	if c.Host == "" {
		return errors.New("zoekt Host field is not set")
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/%s", c.Host, method), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Zoekt "+strings.ToTitle(method)),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	cl := &http.Client{Transport: &nethttp.Transport{}}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(respBody)
}

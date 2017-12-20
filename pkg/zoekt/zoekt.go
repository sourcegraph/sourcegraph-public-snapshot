// Package zoekt provides a client to github.com/google/zoekt
package zoekt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

// Client is a zoekt client to the zoekt rest API.
type Client struct {
	// Host is the hostname of the zoekt instance. It can include a port. For
	// example "localhost:6070".
	Host string
}

// Search sends a search request. Read the documentation for SearchRequest and
// SearchResponse.
func (c *Client) Search(ctx context.Context, d SearchRequest) (*SearchResponse, error) {
	if c.Host == "" {
		return nil, errors.New("zoekt Host field is not set")
	}

	data, err := json.Marshal(d)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search zoekt")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/search", c.Host), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "failed to search zoekt")
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Zoekt Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	cl := &http.Client{Transport: &nethttp.Transport{}}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search zoekt")
	}
	defer resp.Body.Close()

	var sr SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search zoekt")
	}

	return &sr, nil
}

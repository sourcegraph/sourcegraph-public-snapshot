package bitbucketcloud

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type PaginatedResultSet struct {
	mu        sync.Mutex
	initial   *url.URL
	pageToken *PageToken
	nodes     []any
	fetch     func(context.Context, *http.Request) (*PageToken, []any, error)
}

// NewPaginatedResultSet instantiates a new result set. This is intended for
// internal use only, and is exported only to simplify testing in other
// packages.
func NewPaginatedResultSet(initial *url.URL, fetch func(context.Context, *http.Request) (*PageToken, []any, error)) *PaginatedResultSet {
	return &PaginatedResultSet{
		initial: initial,
		fetch:   fetch,
	}
}

// All walks the result set, returning all entries as a single slice.
//
// Note that this essentially consumes the result set.
func (rs *PaginatedResultSet) All(ctx context.Context) ([]any, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var nodes []any
	for {
		node, err := rs.next(ctx)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nodes, nil
		}
		nodes = append(nodes, node)
	}
}

// Next returns the next item in the result set, requesting the next page if
// necessary.
//
// If nil, nil is returned, then there are no further results.
func (rs *PaginatedResultSet) Next(ctx context.Context) (any, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return rs.next(ctx)
}

// WithPageLength configures the size of each page that is requested by the
// result set.
//
// This must be invoked before All or Next are first called, otherwise you may
// receive inconsistent results.
func (rs *PaginatedResultSet) WithPageLength(pageLen int) *PaginatedResultSet {
	initial := *rs.initial
	values := initial.Query()
	values.Set("pagelen", strconv.Itoa(pageLen))
	initial.RawQuery = values.Encode()

	return NewPaginatedResultSet(&initial, rs.fetch)
}

func (rs *PaginatedResultSet) reqPage(ctx context.Context) error {
	req, err := rs.nextPageRequest()
	if err != nil {
		return err
	}

	if req == nil {
		// Nothing to do.
		return nil
	}

	pageToken, page, err := rs.fetch(ctx, req)
	if err != nil {
		return err
	}

	rs.pageToken = pageToken
	rs.nodes = append(rs.nodes, page...)
	return nil
}

func (rs *PaginatedResultSet) nextPageRequest() (*http.Request, error) {
	if rs.pageToken != nil {
		if rs.pageToken.Next == "" {
			// No further pages, so do nothing, successfully.
			return nil, nil
		}

		return http.NewRequest("GET", rs.pageToken.Next, nil)
	}

	return http.NewRequest("GET", rs.initial.String(), nil)
}

func (rs *PaginatedResultSet) next(ctx context.Context) (any, error) {
	// Check if we need to request the next page.
	if len(rs.nodes) == 0 {
		if err := rs.reqPage(ctx); err != nil {
			return nil, err
		}
	}

	// If there are still no nodes, then we've reached the end of the result
	// set.
	if len(rs.nodes) == 0 {
		return nil, nil
	}

	node := rs.nodes[0]
	rs.nodes = rs.nodes[1:]
	return node, nil
}

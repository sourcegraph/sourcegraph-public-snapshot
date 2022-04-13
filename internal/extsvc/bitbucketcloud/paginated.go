package bitbucketcloud

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type ResultSet struct {
	client    *Client
	mu        sync.Mutex
	initial   *url.URL
	pageToken *PageToken
	nodes     []interface{}
	fetch     func(context.Context, *http.Request) (*PageToken, []interface{}, error)
}

func newResultSet(c *Client, initial *url.URL, fetch func(context.Context, *http.Request) (*PageToken, []interface{}, error)) *ResultSet {
	return &ResultSet{
		client:  c,
		initial: initial,
		fetch:   fetch,
	}
}

func (rs *ResultSet) WithPageLength(pageLen int) *ResultSet {
	initial := *rs.initial
	values := initial.Query()
	values.Set("pagelen", strconv.Itoa(pageLen))
	initial.RawQuery = values.Encode()

	return newResultSet(rs.client, &initial, rs.fetch)
}

func (rs *ResultSet) reqPage(ctx context.Context) error {
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

func (rs *ResultSet) nextPageRequest() (*http.Request, error) {
	if rs.pageToken != nil {
		if rs.pageToken.Next == "" {
			// No further pages, so do nothing, successfully.
			return nil, nil
		}

		return http.NewRequest("GET", rs.pageToken.Next, nil)
	}

	return http.NewRequest("GET", rs.initial.String(), nil)
}

func (rs *ResultSet) Next(ctx context.Context) (interface{}, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return rs.next(ctx)
}

func (rs *ResultSet) next(ctx context.Context) (interface{}, error) {
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

// Page returns the current page, or requests the next page if the current page
// is empty.
//
// nil, nil is returned if there are no more pages.
func (rs *ResultSet) Page(ctx context.Context) ([]interface{}, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if len(rs.nodes) > 0 {
		nodes := rs.nodes
		rs.nodes = []interface{}{}

		return nodes, nil
	}

	if err := rs.reqPage(ctx); err != nil {
		return nil, err
	}

	if len(rs.nodes) == 0 {
		return nil, nil
	}

	nodes := rs.nodes
	rs.nodes = []interface{}{}

	return nodes, nil
}

// All walks the result set, returning all entries as a single slice.
//
// Note that this essentially consumes the result set.
func (rs *ResultSet) All(ctx context.Context) ([]interface{}, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var nodes []interface{}
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

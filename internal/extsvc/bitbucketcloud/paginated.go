package bitbucketcloud

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type ResultSet[T any] struct {
	client    *Client
	mu        sync.Mutex
	initial   *url.URL
	pageToken *PageToken
	nodes     []T
}

func newResultSet[T any](c *Client, initial *url.URL) *ResultSet[T] {
	return &ResultSet[T]{
		client:  c,
		initial: initial,
	}
}

func (rs *ResultSet[T]) WithPageLength(pageLen int) *ResultSet[T] {
	initial := *rs.initial
	values := initial.Query()
	values.Set("pagelen", strconv.Itoa(pageLen))
	initial.RawQuery = values.Encode()

	return newResultSet[T](rs.client, &initial)
}

func (rs *ResultSet[T]) reqPage(ctx context.Context, req *http.Request) error {
	result := struct {
		*PageToken
		Values []T `json:"values"`
	}{}

	req, err := rs.nextPageRequest()
	if err != nil {
		return err
	}

	if req == nil {
		// Nothing to do.
		return nil
	}

	if err = rs.client.do(ctx, req, &result); err != nil {
		return err
	}

	rs.pageToken = result.PageToken
	rs.nodes = append(rs.nodes, result.Values...)
	return nil
}

func (rs *ResultSet[T]) nextPageRequest() (*http.Request, error) {
	if rs.pageToken != nil {
		if rs.pageToken.Next == "" {
			// No further pages, so do nothing, successfully.
			return nil, nil
		}

		return http.NewRequest("GET", rs.pageToken.Next, nil)
	}

	return http.NewRequest("GET", rs.initial.String(), nil)
}

func (rs *ResultSet[T]) Next(ctx context.Context) (*T, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Check if we need to request the next page.
	if len(rs.nodes) == 0 {
		if err := rs.reqPage(ctx, nil); err != nil {
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
	return &node, nil
}

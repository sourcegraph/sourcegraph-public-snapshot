package database

import (
	"context"
	"reflect"
	"sync"
)

type mockedResponse struct {
	val any
	err error
}

type mockDBClient struct {
	mutex   sync.Mutex
	mocks   map[reflect.Type]*mockedResponse
	history map[reflect.Type][]DBQuery
}

// Creates a new mockDBClient.
//
// The mockDBClient implements the DBClient interface, but provides additional
// methods to mock DBQuery executions and return pre-determined responses
// instead of performing the actual database queries.
func NewMockDBClient() *mockDBClient {
	return &mockDBClient{
		mocks:   make(map[reflect.Type]*mockedResponse),
		history: make(map[reflect.Type][]DBQuery),
	}
}

// Executes a DBQuery by returning the mocked response for the query.
// If not mocked responses are found, nil is returned instead.
func (c *mockDBClient) Execute(ctx context.Context, q DBQuery) (any, error) {
	rt := reflect.TypeOf(q)
	resp, ok := c.mocks[rt]
	c.mutex.Lock()
	c.history[rt] = append(c.history[rt], q)
	c.mutex.Unlock()
	if !ok {
		return nil, nil
	}

	return resp.val, resp.err
}

// Mocks a DBQuery by registering a response and an error that will be
// returned whenever that query is executed.
func (c *mockDBClient) Mock(req DBQuery, resp any, err error) {
	c.mocks[reflect.TypeOf(req)] = &mockedResponse{
		val: resp,
		err: err,
	}
}

// Returns the query history as a list of DBQueries for the provided DBQuery.
func (c *mockDBClient) History(q DBQuery) []DBQuery {
	queryType := reflect.TypeOf(q)
	c.mutex.Lock()
	history := make([]DBQuery, len(c.history[queryType]))
	copy(history, c.history[queryType])
	c.mutex.Unlock()
	return history
}

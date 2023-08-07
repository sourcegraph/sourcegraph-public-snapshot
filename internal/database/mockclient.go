package database

import (
	"context"
	"reflect"
)

type mockedResponse struct {
	val any
	err error
}

type mockDBClient struct {
	mocks   map[reflect.Type]*mockedResponse
	history map[reflect.Type][]DBQuery
}

func NewMockDBClient() *mockDBClient {
	return &mockDBClient{
		mocks:   make(map[reflect.Type]*mockedResponse),
		history: make(map[reflect.Type][]DBQuery),
	}
}

func (c *mockDBClient) Execute(ctx context.Context, q DBQuery) (any, error) {
	rt := reflect.TypeOf(q)
	resp, ok := c.mocks[rt]
	c.history[rt] = append(c.history[rt], q)
	if !ok {
		return nil, nil
	}

	return resp.val, resp.err
}

func (c *mockDBClient) Mock(req DBQuery, resp any, err error) {
	c.mocks[reflect.TypeOf(req)] = &mockedResponse{
		val: resp,
		err: err,
	}
}

func (c *mockDBClient) History(q DBQuery) []DBQuery {
	return c.history[reflect.TypeOf(q)]
}

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
	mocks map[reflect.Type]*mockedResponse
}

func NewMockDBClient() *mockDBClient {
	return &mockDBClient{
		mocks: make(map[reflect.Type]*mockedResponse),
	}
}

func (c *mockDBClient) Execute(ctx context.Context, q DBQuery) (any, error) {
	resp, ok := c.mocks[reflect.TypeOf(q)]
	if !ok {
		panic("No mock found")
	}

	return resp.val, resp.err
}

func (c *mockDBClient) Mock(req DBQuery, resp any, err error) {
	c.mocks[reflect.TypeOf(req)] = &mockedResponse{
		val: resp,
		err: err,
	}
}

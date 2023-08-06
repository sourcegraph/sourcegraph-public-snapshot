package dbclient

import (
	"context"
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type mockResponse struct {
	val any
	err error
}

type mockDBClient struct {
	*basestore.Store
	baseClient DBClient
	mocks      map[reflect.Type]mockResponse
}

func NewMockClient(db basestore.ShareableStore) *mockDBClient {
	return &mockDBClient{
		Store: basestore.NewWithHandle(db.Handle()),
		mocks: make(map[reflect.Type]mockResponse),
	}
}

func NewMockClientWithBase(baseClient DBClient) *mockDBClient {
	return &mockDBClient{
		baseClient: baseClient,
		mocks:      make(map[reflect.Type]mockResponse),
	}
}

func (c *mockDBClient) Mock(query ExecutableQuery, resp any, err error) {
	c.mocks[reflect.TypeOf(query)] = mockResponse{resp, err}
}

func (c *mockDBClient) Execute(ctx context.Context, query ExecutableQuery) (any, error) {
	resp, ok := c.mocks[reflect.TypeOf(query)]
	if !ok {
		if c.baseClient == nil {
			panic("No mock or base client")
		}
		return c.baseClient.Execute(ctx, query)
	}

	if resp.err != nil {
		return nil, resp.err
	}

	return resp.val, nil
}

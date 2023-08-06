package database

import "context"

type mockDBClient struct {
	mocks map[DBQuery]any
}

func NewMockDBClient() *mockDBClient {
	return &mockDBClient{
		mocks: make(map[DBQuery]any),
	}
}

func (c *mockDBClient) Execute(ctx context.Context, q DBQuery) (any, error) {
	resp, ok := c.mocks[q]
	if !ok {
		panic("No mock found")
	}

	return resp, nil
}

func (c *mockDBClient) Mock(req DBQuery, resp any) {
	c.mocks[req] = resp
}

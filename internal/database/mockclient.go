package database

import (
	"context"
	"reflect"
	"sync"
)

type mockedResponse struct {
	val DBCommand
	err error
}

type mockDBClient struct {
	mutex   sync.Mutex
	history map[reflect.Type][]DBCommand
	mocks   map[reflect.Type]*mockedResponse
}

// Creates a new mockDBClient.
//
// The mockDBClient implements the DBClient interface, but provides additional
// methods to mock DBQuery executions and return pre-determined responses
// instead of performing the actual database queries.
func NewMockDBClient() *mockDBClient {
	return &mockDBClient{
		mocks:   make(map[reflect.Type]*mockedResponse),
		history: make(map[reflect.Type][]DBCommand),
	}
}

func (c *mockDBClient) ExecuteCommand(_ context.Context, command DBCommand) error {
	rt := reflect.TypeOf(command)
	rv := reflect.ValueOf(command)
	newValue := reflect.New(rv.Elem().Type())
	newValue.Elem().Set(rv.Elem())
	commandCopy := newValue.Interface().(DBCommand)
	c.mutex.Lock()
	c.history[rt] = append(c.history[rt], commandCopy)
	c.mutex.Unlock()

	resp, ok := c.mocks[rt]
	if !ok {
		return nil
	}

	rv.Elem().Set(reflect.ValueOf(resp.val).Elem())

	return resp.err
}

func (c *mockDBClient) Mock(command DBCommand, err error) {
	c.mocks[reflect.TypeOf(command)] = &mockedResponse{
		val: command,
		err: err,
	}
}

// Returns the query history as a list of DBQueries for the provided DBQuery.
func (c *mockDBClient) History(command DBCommand) []DBCommand {
	ct := reflect.TypeOf(command)
	c.mutex.Lock()
	history := make([]DBCommand, len(c.history[ct]))
	copy(history, c.history[ct])
	c.mutex.Unlock()
	return history
}

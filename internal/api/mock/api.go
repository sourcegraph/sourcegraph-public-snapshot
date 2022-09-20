package mock

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/stretchr/testify/mock"
)

type Client struct {
	mock.Mock
}

func (m *Client) NewQuery(query string) api.Request {
	args := m.Called(query)
	return args.Get(0).(api.Request)
}

func (m *Client) NewRequest(query string, vars map[string]interface{}) api.Request {
	args := m.Called(query, vars)
	return args.Get(0).(api.Request)
}

func (m *Client) NewGzippedRequest(query string, vars map[string]interface{}) api.Request {
	args := m.Called(query, vars)
	return args.Get(0).(api.Request)
}

func (m *Client) NewGzippedQuery(query string) api.Request {
	args := m.Called(query)
	return args.Get(0).(api.Request)
}

func (m *Client) NewHTTPRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	args := m.Called(ctx, method, path, body)
	var obj *http.Request
	if args.Get(0) != nil {
		obj = args.Get(0).(*http.Request)
	}
	return obj, args.Error(1)
}

func (m *Client) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	var obj *http.Response
	if args.Get(0) != nil {
		obj = args.Get(0).(*http.Response)
	}
	return obj, args.Error(1)
}

type Request struct {
	mock.Mock
	Response string
}

func (r *Request) Do(ctx context.Context, result interface{}) (bool, error) {
	args := r.Called(ctx, result)
	if r.Response != "" {
		if err := json.Unmarshal([]byte(r.Response), result); err != nil {
			return false, err
		}
	}
	return args.Bool(0), args.Error(1)
}

func (r *Request) DoRaw(ctx context.Context, result interface{}) (bool, error) {
	args := r.Called(ctx, result)
	if r.Response != "" {
		if err := json.Unmarshal([]byte(r.Response), result); err != nil {
			return false, err
		}
	}
	return args.Bool(0), args.Error(1)
}

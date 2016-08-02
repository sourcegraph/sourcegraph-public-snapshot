package coverage

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
)

// TODO(slimsag): This *really* does not belong here, *jsonrpc2.Client should
// just be made better / incorporate these changes directly.
type betterClient struct {
	c                           *jsonrpc2.Client
	idMu                        sync.Mutex
	id                          uint32
	idRequestAndWaitForResponse uint32
}

func (c *betterClient) Close() error {
	return c.c.Close()
}

type betterRequest struct {
	Method          string
	Params, Results interface{}
}

func (c *betterClient) RequestBatchAndWaitForAllResponses(reqs ...betterRequest) error {
	reqs2 := make([]*jsonrpc2.Request, len(reqs))
	for i, p := range reqs {
		req2, err := c.buildRequest(p.Method, p.Params)
		if err != nil {
			return err
		}
		reqs2[i] = req2
	}
	resp, err := c.c.RequestBatchAndWaitForAllResponses(reqs2...)
	if err != nil {
		return err
	}
	for i, p := range reqs {
		if p.Results == nil {
			continue
		}
		resp := resp[reqs2[i].ID]
		if resp == nil {
			continue
		}
		if resp.Error != nil {
			return resp.Error
		}
		if err := json.Unmarshal(*resp.Result, p.Results); err != nil {
			return err
		}
	}
	return err
}

func (c *betterClient) buildRequest(method string, params interface{}) (*jsonrpc2.Request, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint32(&c.id, 1)
	return &jsonrpc2.Request{
		ID:     fmt.Sprint(id),
		Method: method,
		Params: (*json.RawMessage)(&data),
	}, nil
}

func (c *betterClient) RequestAndWaitForResponse(method string, params, results interface{}) error {
	req, err := c.buildRequest(method, params)
	if err != nil {
		return err
	}
	resp, err := c.c.RequestAndWaitForResponse(req)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	return json.Unmarshal(*resp.Result, results)
}

func (c *betterClient) Request(method string) error {
	return c.c.Request(&jsonrpc2.Request{Method: method})
}

package zap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
)

// Client is a Zap client.
type Client struct {
	c *jsonrpc2.Conn

	refUpdateCallback         func(context.Context, RefUpdateDownstreamParams) error
	refUpdateSymbolicCallback func(context.Context, RefUpdateSymbolicParams) error

	// ShowStatus, if provided, is called synchronously when the
	// status of the zap client changes. It can indicate that the
	// client is operating as expected, or unable to sync, etc. Only
	// the most recent status should be displayed.
	ShowStatus func(ShowStatusParams)

	// ShowMessage, if provided, is called synchronously when there
	// are messages that should be displayed to the user.
	ShowMessage func(ShowMessageParams)

	mu                 sync.Mutex
	closed             bool
	refUpdates         chan RefUpdateDownstreamParams
	startRefUpdateLoop sync.Once
}

// NewClient creates a new Zap client.
func NewClient(ctx context.Context, stream jsonrpc2.ObjectStream, opt ...jsonrpc2.ConnOpt) *Client {
	var c Client
	// We use a synchronous jsonrpc2 handler to ensure that our client
	// callbacks receive messages in the order intended by the server.
	c.refUpdates = make(chan RefUpdateDownstreamParams, 1000)
	c.c = jsonrpc2.NewConn(ctx, stream, jsonrpc2.HandlerWithError(c.handle), opt...)
	return &c
}

// SetRefUpdateCallback sets the function that is called synchronously
// when the client receives a "ref/update" request from the server.
func (c *Client) SetRefUpdateCallback(f func(context.Context, RefUpdateDownstreamParams) error) {
	if c.refUpdateCallback != nil {
		panic("refUpdateCallback is already set")
	}
	c.refUpdateCallback = f
}

// SetRefUpdateSymbolicCallback sets the function that is called
// synchronously when the client receives a "ref/updateSymbolic"
// request from the server.
func (c *Client) SetRefUpdateSymbolicCallback(f func(context.Context, RefUpdateSymbolicParams) error) {
	if c.refUpdateSymbolicCallback != nil {
		panic("refUpdateSymbolicCallback is already set")
	}
	c.refUpdateSymbolicCallback = f
}

func (c *Client) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	switch req.Method {
	case "ref/update":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefUpdateDownstreamParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if c.refUpdateCallback != nil {
			c.startRefUpdateLoop.Do(func() {
				go func() {
					for {
						select {
						case params, ok := <-c.refUpdates:
							if !ok {
								log.Println("info: ref/update loop is shutting down")
								return
							}
							if err := c.refUpdateCallback(context.Background(), params); err != nil {
								log.Println("warning: client ref/update callback returned error:", err)
							}
						}
					}
				}()
			})
			c.mu.Lock()
			if !c.closed {
				c.refUpdates <- params
			}
			c.mu.Unlock()
			return nil, nil
		}
		log.Printf("warning: client received ref/update from server, but no callback is set: %v", params.string(true))
		return nil, nil

	case "ref/updateSymbolic":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params RefUpdateSymbolicParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		if c.refUpdateSymbolicCallback != nil {
			// The order of symbolic refs is not as important as
			// non-symbolic refs, so we can just process them async.
			go func() {
				if err := c.refUpdateSymbolicCallback(ctx, params); err != nil {
					log.Println("warning: client ref/updateSymbolic callback returned error:", err)
				}
			}()
			return nil, nil
		}
		log.Printf("warning: client received ref/updateSymbolic from server, but no callback is set: %v", params.string(true))
		return nil, nil

	case "window/showStatus":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var p ShowStatusParams
		if err := json.Unmarshal(*req.Params, &p); err != nil {
			return nil, err
		}
		if c.ShowStatus != nil {
			c.ShowStatus(p)
		}
		return true, nil

	case "window/showMessage":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var p ShowMessageParams
		if err := json.Unmarshal(*req.Params, &p); err != nil {
			return nil, err
		}
		if c.ShowMessage != nil {
			c.ShowMessage(p)
		}
		return true, nil

	default:
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: fmt.Sprintf("zap client: method not found: %q", req.Method),
		}
	}
}

// Initialize sends the "initialize" request to the server.
func (c *Client) Initialize(ctx context.Context, params InitializeParams) (res *InitializeResult, err error) {
	return res, c.c.Call(ctx, "initialize", params, &res)
}

// RepoInfo sends the "repo/info" request to the server.
func (c *Client) RepoInfo(ctx context.Context, params RepoInfoParams) (info *RepoInfoResult, err error) {
	err = c.c.Call(ctx, "repo/info", params, &info)
	return
}

// RepoConfigure sends the "repo/configure" request to the server.
func (c *Client) RepoConfigure(ctx context.Context, params RepoConfigureParams) (result *RepoConfigureResult, err error) {
	err = c.c.Call(ctx, "repo/configure", params, &result)
	return
}

// RepoWatch sends the "repo/watch" request to the server.
func (c *Client) RepoWatch(ctx context.Context, params RepoWatchParams) error {
	return c.c.Call(ctx, "repo/watch", params, nil)
}

// RefConfigure sends the "ref/configure" request to the server.
func (c *Client) RefConfigure(ctx context.Context, params RefConfigureParams) error {
	return c.c.Call(ctx, "ref/configure", params, nil)
}

// RefUpdate sends the "ref/update" request to the server.
func (c *Client) RefUpdate(ctx context.Context, params RefUpdateUpstreamParams) error {
	return c.c.Call(ctx, "ref/update", params, nil)
}

// RefUpdateSymbolic sends the "ref/updateSymbolic" request to the
// server.
func (c *Client) RefUpdateSymbolic(ctx context.Context, params RefUpdateSymbolicParams) error {
	return c.c.Call(ctx, "ref/updateSymbolic", params, nil)
}

// RefInfo sends the "ref/info" request to the server.
func (c *Client) RefInfo(ctx context.Context, params RefIdentifier) (*RefInfoResult, error) {
	var state *RefInfoResult
	if err := c.c.Call(ctx, "ref/info", params, &state); err != nil {
		return nil, err
	}
	return state, nil
}

// RefList sends the "ref/list" request to the server.
func (c *Client) RefList(ctx context.Context) ([]RefInfo, error) {
	var r []RefInfo
	if err := c.c.Call(ctx, "ref/list", nil, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// WorkspaceStatus sends the "workspace/status" request to the server.
func (c *Client) WorkspaceStatus(ctx context.Context, params WorkspaceStatusParams) (status *ShowStatusParams, err error) {
	err = c.c.Call(ctx, "workspace/status", params, &status)
	return
}

// WorkspaceAdd sends the "workspace/add" request to the server.
func (c *Client) WorkspaceAdd(ctx context.Context, params WorkspaceAddParams) error {
	return c.c.Call(ctx, "workspace/add", params, nil)
}

// WorkspaceRemove sends the "workspace/remove" request to the server.
func (c *Client) WorkspaceRemove(ctx context.Context, params WorkspaceRemoveParams) error {
	return c.c.Call(ctx, "workspace/remove", params, nil)
}

// WorkspaceCheckout sends the "workspace/checkout" request to the server.
func (c *Client) WorkspaceCheckout(ctx context.Context, params WorkspaceCheckoutParams) error {
	return c.c.Call(ctx, "workspace/checkout", params, nil)
}

// WorkspaceWillSaveFile sends the "workspace/willSaveFile" request to
// the server.
func (c *Client) WorkspaceWillSaveFile(ctx context.Context, params WorkspaceWillSaveFileParams) error {
	return c.c.Call(ctx, "workspace/willSaveFile", params, nil)
}

// Ping sends the "ping" request to the server.
func (c *Client) Ping(ctx context.Context) error {
	return c.c.Call(ctx, "ping", nil, nil)
}

// Wait waits until the underlying connection is closed.
func (c *Client) Wait() {
	<-c.c.DisconnectNotify()
}

// DisconnectNotify returns a channel that is closed when the client
// or its peer disconnects.
func (c *Client) DisconnectNotify() <-chan struct{} {
	return c.c.DisconnectNotify()
}

// Close closes the client's connection.
func (c *Client) Close() error {
	c.mu.Lock()
	if !c.closed {
		c.closed = true
		close(c.refUpdates)
	}
	c.mu.Unlock()
	return c.c.Close()
}

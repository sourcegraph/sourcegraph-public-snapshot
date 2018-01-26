// (experimental)
package inspector

import (
	"github.com/neelance/cdp-go/rpc"
)

// (experimental)
type Client struct {
	*rpc.Client
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables inspector domain notifications.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Inspector.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables inspector domain notifications.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Inspector.disable", r.opts, nil)
}

func init() {
	rpc.EventTypes["Inspector.detached"] = func() interface{} { return new(DetachedEvent) }
	rpc.EventTypes["Inspector.targetCrashed"] = func() interface{} { return new(TargetCrashedEvent) }
}

// Fired when remote debugging connection is about to be terminated. Contains detach reason.
type DetachedEvent struct {
	// The reason why connection has been terminated.
	Reason string `json:"reason"`
}

// Fired when debugging target has crashed
type TargetCrashedEvent struct {
}

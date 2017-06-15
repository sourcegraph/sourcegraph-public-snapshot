// (experimental)
package memory

import (
	"github.com/neelance/cdp-go/rpc"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Memory pressure level.

type PressureLevel string

type GetDOMCountersRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) GetDOMCounters() *GetDOMCountersRequest {
	return &GetDOMCountersRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetDOMCountersResult struct {
	Documents int `json:"documents"`

	Nodes int `json:"nodes"`

	JsEventListeners int `json:"jsEventListeners"`
}

func (r *GetDOMCountersRequest) Do() (*GetDOMCountersResult, error) {
	var result GetDOMCountersResult
	err := r.client.Call("Memory.getDOMCounters", r.opts, &result)
	return &result, err
}

type SetPressureNotificationsSuppressedRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enable/disable suppressing memory pressure notifications in all processes.
func (d *Client) SetPressureNotificationsSuppressed() *SetPressureNotificationsSuppressedRequest {
	return &SetPressureNotificationsSuppressedRequest{opts: make(map[string]interface{}), client: d.Client}
}

// If true, memory pressure notifications will be suppressed.
func (r *SetPressureNotificationsSuppressedRequest) Suppressed(v bool) *SetPressureNotificationsSuppressedRequest {
	r.opts["suppressed"] = v
	return r
}

func (r *SetPressureNotificationsSuppressedRequest) Do() error {
	return r.client.Call("Memory.setPressureNotificationsSuppressed", r.opts, nil)
}

type SimulatePressureNotificationRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Simulate a memory pressure notification in all processes.
func (d *Client) SimulatePressureNotification() *SimulatePressureNotificationRequest {
	return &SimulatePressureNotificationRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Memory pressure level of the notification.
func (r *SimulatePressureNotificationRequest) Level(v PressureLevel) *SimulatePressureNotificationRequest {
	r.opts["level"] = v
	return r
}

func (r *SimulatePressureNotificationRequest) Do() error {
	return r.client.Call("Memory.simulatePressureNotification", r.opts, nil)
}

func init() {
}

// (experimental)
package deviceorientation

import (
	"github.com/neelance/cdp-go/rpc"
)

// (experimental)
type Client struct {
	*rpc.Client
}

type SetDeviceOrientationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Overrides the Device Orientation.
func (d *Client) SetDeviceOrientationOverride() *SetDeviceOrientationOverrideRequest {
	return &SetDeviceOrientationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Mock alpha
func (r *SetDeviceOrientationOverrideRequest) Alpha(v float64) *SetDeviceOrientationOverrideRequest {
	r.opts["alpha"] = v
	return r
}

// Mock beta
func (r *SetDeviceOrientationOverrideRequest) Beta(v float64) *SetDeviceOrientationOverrideRequest {
	r.opts["beta"] = v
	return r
}

// Mock gamma
func (r *SetDeviceOrientationOverrideRequest) Gamma(v float64) *SetDeviceOrientationOverrideRequest {
	r.opts["gamma"] = v
	return r
}

func (r *SetDeviceOrientationOverrideRequest) Do() error {
	return r.client.Call("DeviceOrientation.setDeviceOrientationOverride", r.opts, nil)
}

type ClearDeviceOrientationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears the overridden Device Orientation.
func (d *Client) ClearDeviceOrientationOverride() *ClearDeviceOrientationOverrideRequest {
	return &ClearDeviceOrientationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearDeviceOrientationOverrideRequest) Do() error {
	return r.client.Call("DeviceOrientation.clearDeviceOrientationOverride", r.opts, nil)
}

func init() {
}

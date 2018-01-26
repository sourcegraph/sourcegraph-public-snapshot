// This domain emulates different environments for the page.
package emulation

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/dom"
)

// This domain emulates different environments for the page.
type Client struct {
	*rpc.Client
}

// Screen orientation.

type ScreenOrientation struct {
	// Orientation type.
	Type string `json:"type"`

	// Orientation angle.
	Angle int `json:"angle"`
}

// advance: If the scheduler runs out of immediate work, the virtual time base may fast forward to allow the next delayed task (if any) to run; pause: The virtual time base may not advance; pauseIfNetworkFetchesPending: The virtual time base may not advance if there are any pending resource fetches. (experimental)

type VirtualTimePolicy string

type SetDeviceMetricsOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Overrides the values of device screen dimensions (window.screen.width, window.screen.height, window.innerWidth, window.innerHeight, and "device-width"/"device-height"-related CSS media query results).
func (d *Client) SetDeviceMetricsOverride() *SetDeviceMetricsOverrideRequest {
	return &SetDeviceMetricsOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Overriding width value in pixels (minimum 0, maximum 10000000). 0 disables the override.
func (r *SetDeviceMetricsOverrideRequest) Width(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["width"] = v
	return r
}

// Overriding height value in pixels (minimum 0, maximum 10000000). 0 disables the override.
func (r *SetDeviceMetricsOverrideRequest) Height(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["height"] = v
	return r
}

// Overriding device scale factor value. 0 disables the override.
func (r *SetDeviceMetricsOverrideRequest) DeviceScaleFactor(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["deviceScaleFactor"] = v
	return r
}

// Whether to emulate mobile device. This includes viewport meta tag, overlay scrollbars, text autosizing and more.
func (r *SetDeviceMetricsOverrideRequest) Mobile(v bool) *SetDeviceMetricsOverrideRequest {
	r.opts["mobile"] = v
	return r
}

// Whether a view that exceeds the available browser window area should be scaled down to fit.
func (r *SetDeviceMetricsOverrideRequest) FitWindow(v bool) *SetDeviceMetricsOverrideRequest {
	r.opts["fitWindow"] = v
	return r
}

// Scale to apply to resulting view image. Ignored in |fitWindow| mode. (optional, experimental)
func (r *SetDeviceMetricsOverrideRequest) Scale(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["scale"] = v
	return r
}

// Not used. (optional, experimental)
func (r *SetDeviceMetricsOverrideRequest) OffsetX(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["offsetX"] = v
	return r
}

// Not used. (optional, experimental)
func (r *SetDeviceMetricsOverrideRequest) OffsetY(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["offsetY"] = v
	return r
}

// Overriding screen width value in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional, experimental)
func (r *SetDeviceMetricsOverrideRequest) ScreenWidth(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["screenWidth"] = v
	return r
}

// Overriding screen height value in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional, experimental)
func (r *SetDeviceMetricsOverrideRequest) ScreenHeight(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["screenHeight"] = v
	return r
}

// Overriding view X position on screen in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional, experimental)
func (r *SetDeviceMetricsOverrideRequest) PositionX(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["positionX"] = v
	return r
}

// Overriding view Y position on screen in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional, experimental)
func (r *SetDeviceMetricsOverrideRequest) PositionY(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["positionY"] = v
	return r
}

// Screen orientation override. (optional)
func (r *SetDeviceMetricsOverrideRequest) ScreenOrientation(v *ScreenOrientation) *SetDeviceMetricsOverrideRequest {
	r.opts["screenOrientation"] = v
	return r
}

func (r *SetDeviceMetricsOverrideRequest) Do() error {
	return r.client.Call("Emulation.setDeviceMetricsOverride", r.opts, nil)
}

type ClearDeviceMetricsOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears the overriden device metrics.
func (d *Client) ClearDeviceMetricsOverride() *ClearDeviceMetricsOverrideRequest {
	return &ClearDeviceMetricsOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearDeviceMetricsOverrideRequest) Do() error {
	return r.client.Call("Emulation.clearDeviceMetricsOverride", r.opts, nil)
}

type ForceViewportRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Overrides the visible area of the page. The change is hidden from the page, i.e. the observable scroll position and page scale does not change. In effect, the command moves the specified area of the page into the top-left corner of the frame. (experimental)
func (d *Client) ForceViewport() *ForceViewportRequest {
	return &ForceViewportRequest{opts: make(map[string]interface{}), client: d.Client}
}

// X coordinate of top-left corner of the area (CSS pixels).
func (r *ForceViewportRequest) X(v float64) *ForceViewportRequest {
	r.opts["x"] = v
	return r
}

// Y coordinate of top-left corner of the area (CSS pixels).
func (r *ForceViewportRequest) Y(v float64) *ForceViewportRequest {
	r.opts["y"] = v
	return r
}

// Scale to apply to the area (relative to a page scale of 1.0).
func (r *ForceViewportRequest) Scale(v float64) *ForceViewportRequest {
	r.opts["scale"] = v
	return r
}

func (r *ForceViewportRequest) Do() error {
	return r.client.Call("Emulation.forceViewport", r.opts, nil)
}

type ResetViewportRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Resets the visible area of the page to the original viewport, undoing any effects of the <code>forceViewport</code> command. (experimental)
func (d *Client) ResetViewport() *ResetViewportRequest {
	return &ResetViewportRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ResetViewportRequest) Do() error {
	return r.client.Call("Emulation.resetViewport", r.opts, nil)
}

type ResetPageScaleFactorRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests that page scale factor is reset to initial values. (experimental)
func (d *Client) ResetPageScaleFactor() *ResetPageScaleFactorRequest {
	return &ResetPageScaleFactorRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ResetPageScaleFactorRequest) Do() error {
	return r.client.Call("Emulation.resetPageScaleFactor", r.opts, nil)
}

type SetPageScaleFactorRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets a specified page scale factor. (experimental)
func (d *Client) SetPageScaleFactor() *SetPageScaleFactorRequest {
	return &SetPageScaleFactorRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Page scale factor.
func (r *SetPageScaleFactorRequest) PageScaleFactor(v float64) *SetPageScaleFactorRequest {
	r.opts["pageScaleFactor"] = v
	return r
}

func (r *SetPageScaleFactorRequest) Do() error {
	return r.client.Call("Emulation.setPageScaleFactor", r.opts, nil)
}

type SetVisibleSizeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Resizes the frame/viewport of the page. Note that this does not affect the frame's container (e.g. browser window). Can be used to produce screenshots of the specified size. Not supported on Android. (experimental)
func (d *Client) SetVisibleSize() *SetVisibleSizeRequest {
	return &SetVisibleSizeRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Frame width (DIP).
func (r *SetVisibleSizeRequest) Width(v int) *SetVisibleSizeRequest {
	r.opts["width"] = v
	return r
}

// Frame height (DIP).
func (r *SetVisibleSizeRequest) Height(v int) *SetVisibleSizeRequest {
	r.opts["height"] = v
	return r
}

func (r *SetVisibleSizeRequest) Do() error {
	return r.client.Call("Emulation.setVisibleSize", r.opts, nil)
}

type SetScriptExecutionDisabledRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Switches script execution in the page. (experimental)
func (d *Client) SetScriptExecutionDisabled() *SetScriptExecutionDisabledRequest {
	return &SetScriptExecutionDisabledRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether script execution should be disabled in the page.
func (r *SetScriptExecutionDisabledRequest) Value(v bool) *SetScriptExecutionDisabledRequest {
	r.opts["value"] = v
	return r
}

func (r *SetScriptExecutionDisabledRequest) Do() error {
	return r.client.Call("Emulation.setScriptExecutionDisabled", r.opts, nil)
}

type SetGeolocationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Overrides the Geolocation Position or Error. Omitting any of the parameters emulates position unavailable. (experimental)
func (d *Client) SetGeolocationOverride() *SetGeolocationOverrideRequest {
	return &SetGeolocationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Mock latitude (optional)
func (r *SetGeolocationOverrideRequest) Latitude(v float64) *SetGeolocationOverrideRequest {
	r.opts["latitude"] = v
	return r
}

// Mock longitude (optional)
func (r *SetGeolocationOverrideRequest) Longitude(v float64) *SetGeolocationOverrideRequest {
	r.opts["longitude"] = v
	return r
}

// Mock accuracy (optional)
func (r *SetGeolocationOverrideRequest) Accuracy(v float64) *SetGeolocationOverrideRequest {
	r.opts["accuracy"] = v
	return r
}

func (r *SetGeolocationOverrideRequest) Do() error {
	return r.client.Call("Emulation.setGeolocationOverride", r.opts, nil)
}

type ClearGeolocationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears the overriden Geolocation Position and Error. (experimental)
func (d *Client) ClearGeolocationOverride() *ClearGeolocationOverrideRequest {
	return &ClearGeolocationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearGeolocationOverrideRequest) Do() error {
	return r.client.Call("Emulation.clearGeolocationOverride", r.opts, nil)
}

type SetTouchEmulationEnabledRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Toggles mouse event-based touch event emulation.
func (d *Client) SetTouchEmulationEnabled() *SetTouchEmulationEnabledRequest {
	return &SetTouchEmulationEnabledRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether the touch event emulation should be enabled.
func (r *SetTouchEmulationEnabledRequest) Enabled(v bool) *SetTouchEmulationEnabledRequest {
	r.opts["enabled"] = v
	return r
}

// Touch/gesture events configuration. Default: current platform. (optional)
func (r *SetTouchEmulationEnabledRequest) Configuration(v string) *SetTouchEmulationEnabledRequest {
	r.opts["configuration"] = v
	return r
}

func (r *SetTouchEmulationEnabledRequest) Do() error {
	return r.client.Call("Emulation.setTouchEmulationEnabled", r.opts, nil)
}

type SetEmulatedMediaRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Emulates the given media for CSS media queries.
func (d *Client) SetEmulatedMedia() *SetEmulatedMediaRequest {
	return &SetEmulatedMediaRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Media type to emulate. Empty string disables the override.
func (r *SetEmulatedMediaRequest) Media(v string) *SetEmulatedMediaRequest {
	r.opts["media"] = v
	return r
}

func (r *SetEmulatedMediaRequest) Do() error {
	return r.client.Call("Emulation.setEmulatedMedia", r.opts, nil)
}

type SetCPUThrottlingRateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables CPU throttling to emulate slow CPUs. (experimental)
func (d *Client) SetCPUThrottlingRate() *SetCPUThrottlingRateRequest {
	return &SetCPUThrottlingRateRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Throttling rate as a slowdown factor (1 is no throttle, 2 is 2x slowdown, etc).
func (r *SetCPUThrottlingRateRequest) Rate(v float64) *SetCPUThrottlingRateRequest {
	r.opts["rate"] = v
	return r
}

func (r *SetCPUThrottlingRateRequest) Do() error {
	return r.client.Call("Emulation.setCPUThrottlingRate", r.opts, nil)
}

type CanEmulateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Tells whether emulation is supported. (experimental)
func (d *Client) CanEmulate() *CanEmulateRequest {
	return &CanEmulateRequest{opts: make(map[string]interface{}), client: d.Client}
}

type CanEmulateResult struct {
	// True if emulation is supported.
	Result bool `json:"result"`
}

func (r *CanEmulateRequest) Do() (*CanEmulateResult, error) {
	var result CanEmulateResult
	err := r.client.Call("Emulation.canEmulate", r.opts, &result)
	return &result, err
}

type SetVirtualTimePolicyRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Turns on virtual time for all frames (replacing real-time with a synthetic time source) and sets the current virtual time policy.  Note this supersedes any previous time budget. (experimental)
func (d *Client) SetVirtualTimePolicy() *SetVirtualTimePolicyRequest {
	return &SetVirtualTimePolicyRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetVirtualTimePolicyRequest) Policy(v VirtualTimePolicy) *SetVirtualTimePolicyRequest {
	r.opts["policy"] = v
	return r
}

// If set, after this many virtual milliseconds have elapsed virtual time will be paused and a virtualTimeBudgetExpired event is sent. (optional)
func (r *SetVirtualTimePolicyRequest) Budget(v int) *SetVirtualTimePolicyRequest {
	r.opts["budget"] = v
	return r
}

func (r *SetVirtualTimePolicyRequest) Do() error {
	return r.client.Call("Emulation.setVirtualTimePolicy", r.opts, nil)
}

type SetDefaultBackgroundColorOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets or clears an override of the default background color of the frame. This override is used if the content does not specify one. (experimental)
func (d *Client) SetDefaultBackgroundColorOverride() *SetDefaultBackgroundColorOverrideRequest {
	return &SetDefaultBackgroundColorOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// RGBA of the default background color. If not specified, any existing override will be cleared. (optional)
func (r *SetDefaultBackgroundColorOverrideRequest) Color(v *dom.RGBA) *SetDefaultBackgroundColorOverrideRequest {
	r.opts["color"] = v
	return r
}

func (r *SetDefaultBackgroundColorOverrideRequest) Do() error {
	return r.client.Call("Emulation.setDefaultBackgroundColorOverride", r.opts, nil)
}

func init() {
	rpc.EventTypes["Emulation.virtualTimeBudgetExpired"] = func() interface{} { return new(VirtualTimeBudgetExpiredEvent) }
}

// Notification sent after the virual time budget for the current VirtualTimePolicy has run out. (experimental)
type VirtualTimeBudgetExpiredEvent struct {
}

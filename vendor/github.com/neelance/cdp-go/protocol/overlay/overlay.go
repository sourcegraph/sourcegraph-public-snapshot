// This domain provides various functionality related to drawing atop the inspected page. (experimental)
package overlay

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/dom"
	"github.com/neelance/cdp-go/protocol/runtime"
)

// This domain provides various functionality related to drawing atop the inspected page. (experimental)
type Client struct {
	*rpc.Client
}

// Configuration data for the highlighting of page elements.

type HighlightConfig struct {
	// Whether the node info tooltip should be shown (default: false). (optional)
	ShowInfo bool `json:"showInfo,omitempty"`

	// Whether the rulers should be shown (default: false). (optional)
	ShowRulers bool `json:"showRulers,omitempty"`

	// Whether the extension lines from node to the rulers should be shown (default: false). (optional)
	ShowExtensionLines bool `json:"showExtensionLines,omitempty"`

	// (optional)
	DisplayAsMaterial bool `json:"displayAsMaterial,omitempty"`

	// The content box highlight fill color (default: transparent). (optional)
	ContentColor *dom.RGBA `json:"contentColor,omitempty"`

	// The padding highlight fill color (default: transparent). (optional)
	PaddingColor *dom.RGBA `json:"paddingColor,omitempty"`

	// The border highlight fill color (default: transparent). (optional)
	BorderColor *dom.RGBA `json:"borderColor,omitempty"`

	// The margin highlight fill color (default: transparent). (optional)
	MarginColor *dom.RGBA `json:"marginColor,omitempty"`

	// The event target element highlight fill color (default: transparent). (optional)
	EventTargetColor *dom.RGBA `json:"eventTargetColor,omitempty"`

	// The shape outside fill color (default: transparent). (optional)
	ShapeColor *dom.RGBA `json:"shapeColor,omitempty"`

	// The shape margin fill color (default: transparent). (optional)
	ShapeMarginColor *dom.RGBA `json:"shapeMarginColor,omitempty"`

	// Selectors to highlight relevant nodes. (optional)
	SelectorList string `json:"selectorList,omitempty"`
}

type InspectMode string

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables domain notifications.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Overlay.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables domain notifications.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Overlay.disable", r.opts, nil)
}

type SetShowPaintRectsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests that backend shows paint rectangles
func (d *Client) SetShowPaintRects() *SetShowPaintRectsRequest {
	return &SetShowPaintRectsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// True for showing paint rectangles
func (r *SetShowPaintRectsRequest) Result(v bool) *SetShowPaintRectsRequest {
	r.opts["result"] = v
	return r
}

func (r *SetShowPaintRectsRequest) Do() error {
	return r.client.Call("Overlay.setShowPaintRects", r.opts, nil)
}

type SetShowDebugBordersRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests that backend shows debug borders on layers
func (d *Client) SetShowDebugBorders() *SetShowDebugBordersRequest {
	return &SetShowDebugBordersRequest{opts: make(map[string]interface{}), client: d.Client}
}

// True for showing debug borders
func (r *SetShowDebugBordersRequest) Show(v bool) *SetShowDebugBordersRequest {
	r.opts["show"] = v
	return r
}

func (r *SetShowDebugBordersRequest) Do() error {
	return r.client.Call("Overlay.setShowDebugBorders", r.opts, nil)
}

type SetShowFPSCounterRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests that backend shows the FPS counter
func (d *Client) SetShowFPSCounter() *SetShowFPSCounterRequest {
	return &SetShowFPSCounterRequest{opts: make(map[string]interface{}), client: d.Client}
}

// True for showing the FPS counter
func (r *SetShowFPSCounterRequest) Show(v bool) *SetShowFPSCounterRequest {
	r.opts["show"] = v
	return r
}

func (r *SetShowFPSCounterRequest) Do() error {
	return r.client.Call("Overlay.setShowFPSCounter", r.opts, nil)
}

type SetShowScrollBottleneckRectsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests that backend shows scroll bottleneck rects
func (d *Client) SetShowScrollBottleneckRects() *SetShowScrollBottleneckRectsRequest {
	return &SetShowScrollBottleneckRectsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// True for showing scroll bottleneck rects
func (r *SetShowScrollBottleneckRectsRequest) Show(v bool) *SetShowScrollBottleneckRectsRequest {
	r.opts["show"] = v
	return r
}

func (r *SetShowScrollBottleneckRectsRequest) Do() error {
	return r.client.Call("Overlay.setShowScrollBottleneckRects", r.opts, nil)
}

type SetShowViewportSizeOnResizeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Paints viewport size upon main frame resize.
func (d *Client) SetShowViewportSizeOnResize() *SetShowViewportSizeOnResizeRequest {
	return &SetShowViewportSizeOnResizeRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether to paint size or not.
func (r *SetShowViewportSizeOnResizeRequest) Show(v bool) *SetShowViewportSizeOnResizeRequest {
	r.opts["show"] = v
	return r
}

func (r *SetShowViewportSizeOnResizeRequest) Do() error {
	return r.client.Call("Overlay.setShowViewportSizeOnResize", r.opts, nil)
}

type SetPausedInDebuggerMessageRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) SetPausedInDebuggerMessage() *SetPausedInDebuggerMessageRequest {
	return &SetPausedInDebuggerMessageRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The message to display, also triggers resume and step over controls. (optional)
func (r *SetPausedInDebuggerMessageRequest) Message(v string) *SetPausedInDebuggerMessageRequest {
	r.opts["message"] = v
	return r
}

func (r *SetPausedInDebuggerMessageRequest) Do() error {
	return r.client.Call("Overlay.setPausedInDebuggerMessage", r.opts, nil)
}

type SetSuspendedRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) SetSuspended() *SetSuspendedRequest {
	return &SetSuspendedRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether overlay should be suspended and not consume any resources until resumed.
func (r *SetSuspendedRequest) Suspended(v bool) *SetSuspendedRequest {
	r.opts["suspended"] = v
	return r
}

func (r *SetSuspendedRequest) Do() error {
	return r.client.Call("Overlay.setSuspended", r.opts, nil)
}

type SetInspectModeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enters the 'inspect' mode. In this mode, elements that user is hovering over are highlighted. Backend then generates 'inspectNodeRequested' event upon element selection.
func (d *Client) SetInspectMode() *SetInspectModeRequest {
	return &SetInspectModeRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Set an inspection mode.
func (r *SetInspectModeRequest) Mode(v InspectMode) *SetInspectModeRequest {
	r.opts["mode"] = v
	return r
}

// A descriptor for the highlight appearance of hovered-over nodes. May be omitted if <code>enabled == false</code>. (optional)
func (r *SetInspectModeRequest) HighlightConfig(v *HighlightConfig) *SetInspectModeRequest {
	r.opts["highlightConfig"] = v
	return r
}

func (r *SetInspectModeRequest) Do() error {
	return r.client.Call("Overlay.setInspectMode", r.opts, nil)
}

type HighlightRectRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Highlights given rectangle. Coordinates are absolute with respect to the main frame viewport.
func (d *Client) HighlightRect() *HighlightRectRequest {
	return &HighlightRectRequest{opts: make(map[string]interface{}), client: d.Client}
}

// X coordinate
func (r *HighlightRectRequest) X(v int) *HighlightRectRequest {
	r.opts["x"] = v
	return r
}

// Y coordinate
func (r *HighlightRectRequest) Y(v int) *HighlightRectRequest {
	r.opts["y"] = v
	return r
}

// Rectangle width
func (r *HighlightRectRequest) Width(v int) *HighlightRectRequest {
	r.opts["width"] = v
	return r
}

// Rectangle height
func (r *HighlightRectRequest) Height(v int) *HighlightRectRequest {
	r.opts["height"] = v
	return r
}

// The highlight fill color (default: transparent). (optional)
func (r *HighlightRectRequest) Color(v *dom.RGBA) *HighlightRectRequest {
	r.opts["color"] = v
	return r
}

// The highlight outline color (default: transparent). (optional)
func (r *HighlightRectRequest) OutlineColor(v *dom.RGBA) *HighlightRectRequest {
	r.opts["outlineColor"] = v
	return r
}

func (r *HighlightRectRequest) Do() error {
	return r.client.Call("Overlay.highlightRect", r.opts, nil)
}

type HighlightQuadRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Highlights given quad. Coordinates are absolute with respect to the main frame viewport.
func (d *Client) HighlightQuad() *HighlightQuadRequest {
	return &HighlightQuadRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Quad to highlight
func (r *HighlightQuadRequest) Quad(v dom.Quad) *HighlightQuadRequest {
	r.opts["quad"] = v
	return r
}

// The highlight fill color (default: transparent). (optional)
func (r *HighlightQuadRequest) Color(v *dom.RGBA) *HighlightQuadRequest {
	r.opts["color"] = v
	return r
}

// The highlight outline color (default: transparent). (optional)
func (r *HighlightQuadRequest) OutlineColor(v *dom.RGBA) *HighlightQuadRequest {
	r.opts["outlineColor"] = v
	return r
}

func (r *HighlightQuadRequest) Do() error {
	return r.client.Call("Overlay.highlightQuad", r.opts, nil)
}

type HighlightNodeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Highlights DOM node with given id or with the given JavaScript object wrapper. Either nodeId or objectId must be specified.
func (d *Client) HighlightNode() *HighlightNodeRequest {
	return &HighlightNodeRequest{opts: make(map[string]interface{}), client: d.Client}
}

// A descriptor for the highlight appearance.
func (r *HighlightNodeRequest) HighlightConfig(v *HighlightConfig) *HighlightNodeRequest {
	r.opts["highlightConfig"] = v
	return r
}

// Identifier of the node to highlight. (optional)
func (r *HighlightNodeRequest) NodeId(v dom.NodeId) *HighlightNodeRequest {
	r.opts["nodeId"] = v
	return r
}

// Identifier of the backend node to highlight. (optional)
func (r *HighlightNodeRequest) BackendNodeId(v dom.BackendNodeId) *HighlightNodeRequest {
	r.opts["backendNodeId"] = v
	return r
}

// JavaScript object id of the node to be highlighted. (optional)
func (r *HighlightNodeRequest) ObjectId(v runtime.RemoteObjectId) *HighlightNodeRequest {
	r.opts["objectId"] = v
	return r
}

func (r *HighlightNodeRequest) Do() error {
	return r.client.Call("Overlay.highlightNode", r.opts, nil)
}

type HighlightFrameRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Highlights owner element of the frame with given id.
func (d *Client) HighlightFrame() *HighlightFrameRequest {
	return &HighlightFrameRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the frame to highlight.
func (r *HighlightFrameRequest) FrameId(v string) *HighlightFrameRequest {
	r.opts["frameId"] = v
	return r
}

// The content box highlight fill color (default: transparent). (optional)
func (r *HighlightFrameRequest) ContentColor(v *dom.RGBA) *HighlightFrameRequest {
	r.opts["contentColor"] = v
	return r
}

// The content box highlight outline color (default: transparent). (optional)
func (r *HighlightFrameRequest) ContentOutlineColor(v *dom.RGBA) *HighlightFrameRequest {
	r.opts["contentOutlineColor"] = v
	return r
}

func (r *HighlightFrameRequest) Do() error {
	return r.client.Call("Overlay.highlightFrame", r.opts, nil)
}

type HideHighlightRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Hides any highlight.
func (d *Client) HideHighlight() *HideHighlightRequest {
	return &HideHighlightRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *HideHighlightRequest) Do() error {
	return r.client.Call("Overlay.hideHighlight", r.opts, nil)
}

type GetHighlightObjectForTestRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// For testing.
func (d *Client) GetHighlightObjectForTest() *GetHighlightObjectForTestRequest {
	return &GetHighlightObjectForTestRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Id of the node to get highlight object for.
func (r *GetHighlightObjectForTestRequest) NodeId(v dom.NodeId) *GetHighlightObjectForTestRequest {
	r.opts["nodeId"] = v
	return r
}

type GetHighlightObjectForTestResult struct {
	// Highlight data for the node.
	Highlight interface{} `json:"highlight"`
}

func (r *GetHighlightObjectForTestRequest) Do() (*GetHighlightObjectForTestResult, error) {
	var result GetHighlightObjectForTestResult
	err := r.client.Call("Overlay.getHighlightObjectForTest", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["Overlay.nodeHighlightRequested"] = func() interface{} { return new(NodeHighlightRequestedEvent) }
	rpc.EventTypes["Overlay.inspectNodeRequested"] = func() interface{} { return new(InspectNodeRequestedEvent) }
}

// Fired when the node should be highlighted. This happens after call to <code>setInspectMode</code>.
type NodeHighlightRequestedEvent struct {
	NodeId dom.NodeId `json:"nodeId"`
}

// Fired when the node should be inspected. This happens after call to <code>setInspectMode</code> or when user manually inspects an element.
type InspectNodeRequestedEvent struct {
	// Id of the node to inspect.
	BackendNodeId dom.BackendNodeId `json:"backendNodeId"`
}

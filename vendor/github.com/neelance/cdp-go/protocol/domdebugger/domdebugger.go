// DOM debugging allows setting breakpoints on particular DOM operations and events. JavaScript execution will stop on these operations as if there was a regular breakpoint set.
package domdebugger

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/dom"
	"github.com/neelance/cdp-go/protocol/runtime"
)

// DOM debugging allows setting breakpoints on particular DOM operations and events. JavaScript execution will stop on these operations as if there was a regular breakpoint set.
type Client struct {
	*rpc.Client
}

// DOM breakpoint type.

type DOMBreakpointType string

// Object event listener. (experimental)

type EventListener struct {
	// <code>EventListener</code>'s type.
	Type string `json:"type"`

	// <code>EventListener</code>'s useCapture.
	UseCapture bool `json:"useCapture"`

	// <code>EventListener</code>'s passive flag.
	Passive bool `json:"passive"`

	// <code>EventListener</code>'s once flag.
	Once bool `json:"once"`

	// Script id of the handler code.
	ScriptId runtime.ScriptId `json:"scriptId"`

	// Line number in the script (0-based).
	LineNumber int `json:"lineNumber"`

	// Column number in the script (0-based).
	ColumnNumber int `json:"columnNumber"`

	// Event handler function value. (optional)
	Handler *runtime.RemoteObject `json:"handler,omitempty"`

	// Event original handler function value. (optional)
	OriginalHandler *runtime.RemoteObject `json:"originalHandler,omitempty"`

	// Node the listener is added to (if any). (optional)
	BackendNodeId dom.BackendNodeId `json:"backendNodeId,omitempty"`
}

type SetDOMBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets breakpoint on particular operation with DOM.
func (d *Client) SetDOMBreakpoint() *SetDOMBreakpointRequest {
	return &SetDOMBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the node to set breakpoint on.
func (r *SetDOMBreakpointRequest) NodeId(v dom.NodeId) *SetDOMBreakpointRequest {
	r.opts["nodeId"] = v
	return r
}

// Type of the operation to stop upon.
func (r *SetDOMBreakpointRequest) Type(v DOMBreakpointType) *SetDOMBreakpointRequest {
	r.opts["type"] = v
	return r
}

func (r *SetDOMBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.setDOMBreakpoint", r.opts, nil)
}

type RemoveDOMBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Removes DOM breakpoint that was set using <code>setDOMBreakpoint</code>.
func (d *Client) RemoveDOMBreakpoint() *RemoveDOMBreakpointRequest {
	return &RemoveDOMBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the node to remove breakpoint from.
func (r *RemoveDOMBreakpointRequest) NodeId(v dom.NodeId) *RemoveDOMBreakpointRequest {
	r.opts["nodeId"] = v
	return r
}

// Type of the breakpoint to remove.
func (r *RemoveDOMBreakpointRequest) Type(v DOMBreakpointType) *RemoveDOMBreakpointRequest {
	r.opts["type"] = v
	return r
}

func (r *RemoveDOMBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.removeDOMBreakpoint", r.opts, nil)
}

type SetEventListenerBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets breakpoint on particular DOM event.
func (d *Client) SetEventListenerBreakpoint() *SetEventListenerBreakpointRequest {
	return &SetEventListenerBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// DOM Event name to stop on (any DOM event will do).
func (r *SetEventListenerBreakpointRequest) EventName(v string) *SetEventListenerBreakpointRequest {
	r.opts["eventName"] = v
	return r
}

// EventTarget interface name to stop on. If equal to <code>"*"</code> or not provided, will stop on any EventTarget. (optional, experimental)
func (r *SetEventListenerBreakpointRequest) TargetName(v string) *SetEventListenerBreakpointRequest {
	r.opts["targetName"] = v
	return r
}

func (r *SetEventListenerBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.setEventListenerBreakpoint", r.opts, nil)
}

type RemoveEventListenerBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Removes breakpoint on particular DOM event.
func (d *Client) RemoveEventListenerBreakpoint() *RemoveEventListenerBreakpointRequest {
	return &RemoveEventListenerBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Event name.
func (r *RemoveEventListenerBreakpointRequest) EventName(v string) *RemoveEventListenerBreakpointRequest {
	r.opts["eventName"] = v
	return r
}

// EventTarget interface name. (optional, experimental)
func (r *RemoveEventListenerBreakpointRequest) TargetName(v string) *RemoveEventListenerBreakpointRequest {
	r.opts["targetName"] = v
	return r
}

func (r *RemoveEventListenerBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.removeEventListenerBreakpoint", r.opts, nil)
}

type SetInstrumentationBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets breakpoint on particular native event. (experimental)
func (d *Client) SetInstrumentationBreakpoint() *SetInstrumentationBreakpointRequest {
	return &SetInstrumentationBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Instrumentation name to stop on.
func (r *SetInstrumentationBreakpointRequest) EventName(v string) *SetInstrumentationBreakpointRequest {
	r.opts["eventName"] = v
	return r
}

func (r *SetInstrumentationBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.setInstrumentationBreakpoint", r.opts, nil)
}

type RemoveInstrumentationBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Removes breakpoint on particular native event. (experimental)
func (d *Client) RemoveInstrumentationBreakpoint() *RemoveInstrumentationBreakpointRequest {
	return &RemoveInstrumentationBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Instrumentation name to stop on.
func (r *RemoveInstrumentationBreakpointRequest) EventName(v string) *RemoveInstrumentationBreakpointRequest {
	r.opts["eventName"] = v
	return r
}

func (r *RemoveInstrumentationBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.removeInstrumentationBreakpoint", r.opts, nil)
}

type SetXHRBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets breakpoint on XMLHttpRequest.
func (d *Client) SetXHRBreakpoint() *SetXHRBreakpointRequest {
	return &SetXHRBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Resource URL substring. All XHRs having this substring in the URL will get stopped upon.
func (r *SetXHRBreakpointRequest) URL(v string) *SetXHRBreakpointRequest {
	r.opts["url"] = v
	return r
}

func (r *SetXHRBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.setXHRBreakpoint", r.opts, nil)
}

type RemoveXHRBreakpointRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Removes breakpoint from XMLHttpRequest.
func (d *Client) RemoveXHRBreakpoint() *RemoveXHRBreakpointRequest {
	return &RemoveXHRBreakpointRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Resource URL substring.
func (r *RemoveXHRBreakpointRequest) URL(v string) *RemoveXHRBreakpointRequest {
	r.opts["url"] = v
	return r
}

func (r *RemoveXHRBreakpointRequest) Do() error {
	return r.client.Call("DOMDebugger.removeXHRBreakpoint", r.opts, nil)
}

type GetEventListenersRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns event listeners of the given object. (experimental)
func (d *Client) GetEventListeners() *GetEventListenersRequest {
	return &GetEventListenersRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the object to return listeners for.
func (r *GetEventListenersRequest) ObjectId(v runtime.RemoteObjectId) *GetEventListenersRequest {
	r.opts["objectId"] = v
	return r
}

// The maximum depth at which Node children should be retrieved, defaults to 1. Use -1 for the entire subtree or provide an integer larger than 0. (optional, experimental)
func (r *GetEventListenersRequest) Depth(v int) *GetEventListenersRequest {
	r.opts["depth"] = v
	return r
}

// Whether or not iframes and shadow roots should be traversed when returning the subtree (default is false). Reports listeners for all contexts if pierce is enabled. (optional, experimental)
func (r *GetEventListenersRequest) Pierce(v bool) *GetEventListenersRequest {
	r.opts["pierce"] = v
	return r
}

type GetEventListenersResult struct {
	// Array of relevant listeners.
	Listeners []*EventListener `json:"listeners"`
}

func (r *GetEventListenersRequest) Do() (*GetEventListenersResult, error) {
	var result GetEventListenersResult
	err := r.client.Call("DOMDebugger.getEventListeners", r.opts, &result)
	return &result, err
}

func init() {
}

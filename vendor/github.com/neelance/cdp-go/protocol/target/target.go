// Supports additional targets discovery and allows to attach to them. (experimental)
package target

import (
	"github.com/neelance/cdp-go/rpc"
)

// Supports additional targets discovery and allows to attach to them. (experimental)
type Client struct {
	*rpc.Client
}

type TargetID string

type BrowserContextID string

type TargetInfo struct {
	TargetId TargetID `json:"targetId"`

	Type string `json:"type"`

	Title string `json:"title"`

	URL string `json:"url"`
}

type RemoteLocation struct {
	Host string `json:"host"`

	Port int `json:"port"`
}

type SetDiscoverTargetsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Controls whether to discover available targets and notify via <code>targetCreated/targetDestroyed</code> events.
func (d *Client) SetDiscoverTargets() *SetDiscoverTargetsRequest {
	return &SetDiscoverTargetsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether to discover available targets.
func (r *SetDiscoverTargetsRequest) Discover(v bool) *SetDiscoverTargetsRequest {
	r.opts["discover"] = v
	return r
}

func (r *SetDiscoverTargetsRequest) Do() error {
	return r.client.Call("Target.setDiscoverTargets", r.opts, nil)
}

type SetAutoAttachRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Controls whether to automatically attach to new targets which are considered to be related to this one. When turned on, attaches to all existing related targets as well. When turned off, automatically detaches from all currently attached targets.
func (d *Client) SetAutoAttach() *SetAutoAttachRequest {
	return &SetAutoAttachRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether to auto-attach to related targets.
func (r *SetAutoAttachRequest) AutoAttach(v bool) *SetAutoAttachRequest {
	r.opts["autoAttach"] = v
	return r
}

// Whether to pause new targets when attaching to them. Use <code>Runtime.runIfWaitingForDebugger</code> to run paused targets.
func (r *SetAutoAttachRequest) WaitForDebuggerOnStart(v bool) *SetAutoAttachRequest {
	r.opts["waitForDebuggerOnStart"] = v
	return r
}

func (r *SetAutoAttachRequest) Do() error {
	return r.client.Call("Target.setAutoAttach", r.opts, nil)
}

type SetAttachToFramesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) SetAttachToFrames() *SetAttachToFramesRequest {
	return &SetAttachToFramesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether to attach to frames.
func (r *SetAttachToFramesRequest) Value(v bool) *SetAttachToFramesRequest {
	r.opts["value"] = v
	return r
}

func (r *SetAttachToFramesRequest) Do() error {
	return r.client.Call("Target.setAttachToFrames", r.opts, nil)
}

type SetRemoteLocationsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables target discovery for the specified locations, when <code>setDiscoverTargets</code> was set to <code>true</code>.
func (d *Client) SetRemoteLocations() *SetRemoteLocationsRequest {
	return &SetRemoteLocationsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// List of remote locations.
func (r *SetRemoteLocationsRequest) Locations(v []*RemoteLocation) *SetRemoteLocationsRequest {
	r.opts["locations"] = v
	return r
}

func (r *SetRemoteLocationsRequest) Do() error {
	return r.client.Call("Target.setRemoteLocations", r.opts, nil)
}

type SendMessageToTargetRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sends protocol message to the target with given id.
func (d *Client) SendMessageToTarget() *SendMessageToTargetRequest {
	return &SendMessageToTargetRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SendMessageToTargetRequest) TargetId(v TargetID) *SendMessageToTargetRequest {
	r.opts["targetId"] = v
	return r
}

func (r *SendMessageToTargetRequest) Message(v string) *SendMessageToTargetRequest {
	r.opts["message"] = v
	return r
}

func (r *SendMessageToTargetRequest) Do() error {
	return r.client.Call("Target.sendMessageToTarget", r.opts, nil)
}

type GetTargetInfoRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns information about a target.
func (d *Client) GetTargetInfo() *GetTargetInfoRequest {
	return &GetTargetInfoRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetTargetInfoRequest) TargetId(v TargetID) *GetTargetInfoRequest {
	r.opts["targetId"] = v
	return r
}

type GetTargetInfoResult struct {
	TargetInfo *TargetInfo `json:"targetInfo"`
}

func (r *GetTargetInfoRequest) Do() (*GetTargetInfoResult, error) {
	var result GetTargetInfoResult
	err := r.client.Call("Target.getTargetInfo", r.opts, &result)
	return &result, err
}

type ActivateTargetRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Activates (focuses) the target.
func (d *Client) ActivateTarget() *ActivateTargetRequest {
	return &ActivateTargetRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ActivateTargetRequest) TargetId(v TargetID) *ActivateTargetRequest {
	r.opts["targetId"] = v
	return r
}

func (r *ActivateTargetRequest) Do() error {
	return r.client.Call("Target.activateTarget", r.opts, nil)
}

type CloseTargetRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Closes the target. If the target is a page that gets closed too.
func (d *Client) CloseTarget() *CloseTargetRequest {
	return &CloseTargetRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *CloseTargetRequest) TargetId(v TargetID) *CloseTargetRequest {
	r.opts["targetId"] = v
	return r
}

type CloseTargetResult struct {
	Success bool `json:"success"`
}

func (r *CloseTargetRequest) Do() (*CloseTargetResult, error) {
	var result CloseTargetResult
	err := r.client.Call("Target.closeTarget", r.opts, &result)
	return &result, err
}

type AttachToTargetRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Attaches to the target with given id.
func (d *Client) AttachToTarget() *AttachToTargetRequest {
	return &AttachToTargetRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *AttachToTargetRequest) TargetId(v TargetID) *AttachToTargetRequest {
	r.opts["targetId"] = v
	return r
}

type AttachToTargetResult struct {
	// Whether attach succeeded.
	Success bool `json:"success"`
}

func (r *AttachToTargetRequest) Do() (*AttachToTargetResult, error) {
	var result AttachToTargetResult
	err := r.client.Call("Target.attachToTarget", r.opts, &result)
	return &result, err
}

type DetachFromTargetRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Detaches from the target with given id.
func (d *Client) DetachFromTarget() *DetachFromTargetRequest {
	return &DetachFromTargetRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DetachFromTargetRequest) TargetId(v TargetID) *DetachFromTargetRequest {
	r.opts["targetId"] = v
	return r
}

func (r *DetachFromTargetRequest) Do() error {
	return r.client.Call("Target.detachFromTarget", r.opts, nil)
}

type CreateBrowserContextRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Creates a new empty BrowserContext. Similar to an incognito profile but you can have more than one.
func (d *Client) CreateBrowserContext() *CreateBrowserContextRequest {
	return &CreateBrowserContextRequest{opts: make(map[string]interface{}), client: d.Client}
}

type CreateBrowserContextResult struct {
	// The id of the context created.
	BrowserContextId BrowserContextID `json:"browserContextId"`
}

func (r *CreateBrowserContextRequest) Do() (*CreateBrowserContextResult, error) {
	var result CreateBrowserContextResult
	err := r.client.Call("Target.createBrowserContext", r.opts, &result)
	return &result, err
}

type DisposeBrowserContextRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Deletes a BrowserContext, will fail of any open page uses it.
func (d *Client) DisposeBrowserContext() *DisposeBrowserContextRequest {
	return &DisposeBrowserContextRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisposeBrowserContextRequest) BrowserContextId(v BrowserContextID) *DisposeBrowserContextRequest {
	r.opts["browserContextId"] = v
	return r
}

type DisposeBrowserContextResult struct {
	Success bool `json:"success"`
}

func (r *DisposeBrowserContextRequest) Do() (*DisposeBrowserContextResult, error) {
	var result DisposeBrowserContextResult
	err := r.client.Call("Target.disposeBrowserContext", r.opts, &result)
	return &result, err
}

type CreateTargetRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Creates a new page.
func (d *Client) CreateTarget() *CreateTargetRequest {
	return &CreateTargetRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The initial URL the page will be navigated to.
func (r *CreateTargetRequest) URL(v string) *CreateTargetRequest {
	r.opts["url"] = v
	return r
}

// Frame width in DIP (headless chrome only). (optional)
func (r *CreateTargetRequest) Width(v int) *CreateTargetRequest {
	r.opts["width"] = v
	return r
}

// Frame height in DIP (headless chrome only). (optional)
func (r *CreateTargetRequest) Height(v int) *CreateTargetRequest {
	r.opts["height"] = v
	return r
}

// The browser context to create the page in (headless chrome only). (optional)
func (r *CreateTargetRequest) BrowserContextId(v BrowserContextID) *CreateTargetRequest {
	r.opts["browserContextId"] = v
	return r
}

type CreateTargetResult struct {
	// The id of the page opened.
	TargetId TargetID `json:"targetId"`
}

func (r *CreateTargetRequest) Do() (*CreateTargetResult, error) {
	var result CreateTargetResult
	err := r.client.Call("Target.createTarget", r.opts, &result)
	return &result, err
}

type GetTargetsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Retrieves a list of available targets.
func (d *Client) GetTargets() *GetTargetsRequest {
	return &GetTargetsRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetTargetsResult struct {
	// The list of targets.
	TargetInfos []*TargetInfo `json:"targetInfos"`
}

func (r *GetTargetsRequest) Do() (*GetTargetsResult, error) {
	var result GetTargetsResult
	err := r.client.Call("Target.getTargets", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["Target.targetCreated"] = func() interface{} { return new(TargetCreatedEvent) }
	rpc.EventTypes["Target.targetDestroyed"] = func() interface{} { return new(TargetDestroyedEvent) }
	rpc.EventTypes["Target.attachedToTarget"] = func() interface{} { return new(AttachedToTargetEvent) }
	rpc.EventTypes["Target.detachedFromTarget"] = func() interface{} { return new(DetachedFromTargetEvent) }
	rpc.EventTypes["Target.receivedMessageFromTarget"] = func() interface{} { return new(ReceivedMessageFromTargetEvent) }
}

// Issued when a possible inspection target is created.
type TargetCreatedEvent struct {
	TargetInfo *TargetInfo `json:"targetInfo"`
}

// Issued when a target is destroyed.
type TargetDestroyedEvent struct {
	TargetId TargetID `json:"targetId"`
}

// Issued when attached to target because of auto-attach or <code>attachToTarget</code> command.
type AttachedToTargetEvent struct {
	TargetInfo *TargetInfo `json:"targetInfo"`

	WaitingForDebugger bool `json:"waitingForDebugger"`
}

// Issued when detached from target for any reason (including <code>detachFromTarget</code> command).
type DetachedFromTargetEvent struct {
	TargetId TargetID `json:"targetId"`
}

// Notifies about new protocol message from attached target.
type ReceivedMessageFromTargetEvent struct {
	TargetId TargetID `json:"targetId"`

	Message string `json:"message"`
}

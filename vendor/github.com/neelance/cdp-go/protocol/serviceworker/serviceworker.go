// (experimental)
package serviceworker

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/target"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// ServiceWorker registration.

type ServiceWorkerRegistration struct {
	RegistrationId string `json:"registrationId"`

	ScopeURL string `json:"scopeURL"`

	IsDeleted bool `json:"isDeleted"`
}

type ServiceWorkerVersionRunningStatus string

type ServiceWorkerVersionStatus string

// ServiceWorker version.

type ServiceWorkerVersion struct {
	VersionId string `json:"versionId"`

	RegistrationId string `json:"registrationId"`

	ScriptURL string `json:"scriptURL"`

	RunningStatus ServiceWorkerVersionRunningStatus `json:"runningStatus"`

	Status ServiceWorkerVersionStatus `json:"status"`

	// The Last-Modified header value of the main script. (optional)
	ScriptLastModified float64 `json:"scriptLastModified,omitempty"`

	// The time at which the response headers of the main script were received from the server.  For cached script it is the last time the cache entry was validated. (optional)
	ScriptResponseTime float64 `json:"scriptResponseTime,omitempty"`

	// (optional)
	ControlledClients []target.TargetID `json:"controlledClients,omitempty"`

	// (optional)
	TargetId target.TargetID `json:"targetId,omitempty"`
}

// ServiceWorker error message.

type ServiceWorkerErrorMessage struct {
	ErrorMessage string `json:"errorMessage"`

	RegistrationId string `json:"registrationId"`

	VersionId string `json:"versionId"`

	SourceURL string `json:"sourceURL"`

	LineNumber int `json:"lineNumber"`

	ColumnNumber int `json:"columnNumber"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("ServiceWorker.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("ServiceWorker.disable", r.opts, nil)
}

type UnregisterRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Unregister() *UnregisterRequest {
	return &UnregisterRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *UnregisterRequest) ScopeURL(v string) *UnregisterRequest {
	r.opts["scopeURL"] = v
	return r
}

func (r *UnregisterRequest) Do() error {
	return r.client.Call("ServiceWorker.unregister", r.opts, nil)
}

type UpdateRegistrationRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) UpdateRegistration() *UpdateRegistrationRequest {
	return &UpdateRegistrationRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *UpdateRegistrationRequest) ScopeURL(v string) *UpdateRegistrationRequest {
	r.opts["scopeURL"] = v
	return r
}

func (r *UpdateRegistrationRequest) Do() error {
	return r.client.Call("ServiceWorker.updateRegistration", r.opts, nil)
}

type StartWorkerRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) StartWorker() *StartWorkerRequest {
	return &StartWorkerRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *StartWorkerRequest) ScopeURL(v string) *StartWorkerRequest {
	r.opts["scopeURL"] = v
	return r
}

func (r *StartWorkerRequest) Do() error {
	return r.client.Call("ServiceWorker.startWorker", r.opts, nil)
}

type SkipWaitingRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) SkipWaiting() *SkipWaitingRequest {
	return &SkipWaitingRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SkipWaitingRequest) ScopeURL(v string) *SkipWaitingRequest {
	r.opts["scopeURL"] = v
	return r
}

func (r *SkipWaitingRequest) Do() error {
	return r.client.Call("ServiceWorker.skipWaiting", r.opts, nil)
}

type StopWorkerRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) StopWorker() *StopWorkerRequest {
	return &StopWorkerRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *StopWorkerRequest) VersionId(v string) *StopWorkerRequest {
	r.opts["versionId"] = v
	return r
}

func (r *StopWorkerRequest) Do() error {
	return r.client.Call("ServiceWorker.stopWorker", r.opts, nil)
}

type InspectWorkerRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) InspectWorker() *InspectWorkerRequest {
	return &InspectWorkerRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *InspectWorkerRequest) VersionId(v string) *InspectWorkerRequest {
	r.opts["versionId"] = v
	return r
}

func (r *InspectWorkerRequest) Do() error {
	return r.client.Call("ServiceWorker.inspectWorker", r.opts, nil)
}

type SetForceUpdateOnPageLoadRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) SetForceUpdateOnPageLoad() *SetForceUpdateOnPageLoadRequest {
	return &SetForceUpdateOnPageLoadRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetForceUpdateOnPageLoadRequest) ForceUpdateOnPageLoad(v bool) *SetForceUpdateOnPageLoadRequest {
	r.opts["forceUpdateOnPageLoad"] = v
	return r
}

func (r *SetForceUpdateOnPageLoadRequest) Do() error {
	return r.client.Call("ServiceWorker.setForceUpdateOnPageLoad", r.opts, nil)
}

type DeliverPushMessageRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) DeliverPushMessage() *DeliverPushMessageRequest {
	return &DeliverPushMessageRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DeliverPushMessageRequest) Origin(v string) *DeliverPushMessageRequest {
	r.opts["origin"] = v
	return r
}

func (r *DeliverPushMessageRequest) RegistrationId(v string) *DeliverPushMessageRequest {
	r.opts["registrationId"] = v
	return r
}

func (r *DeliverPushMessageRequest) Data(v string) *DeliverPushMessageRequest {
	r.opts["data"] = v
	return r
}

func (r *DeliverPushMessageRequest) Do() error {
	return r.client.Call("ServiceWorker.deliverPushMessage", r.opts, nil)
}

type DispatchSyncEventRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) DispatchSyncEvent() *DispatchSyncEventRequest {
	return &DispatchSyncEventRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DispatchSyncEventRequest) Origin(v string) *DispatchSyncEventRequest {
	r.opts["origin"] = v
	return r
}

func (r *DispatchSyncEventRequest) RegistrationId(v string) *DispatchSyncEventRequest {
	r.opts["registrationId"] = v
	return r
}

func (r *DispatchSyncEventRequest) Tag(v string) *DispatchSyncEventRequest {
	r.opts["tag"] = v
	return r
}

func (r *DispatchSyncEventRequest) LastChance(v bool) *DispatchSyncEventRequest {
	r.opts["lastChance"] = v
	return r
}

func (r *DispatchSyncEventRequest) Do() error {
	return r.client.Call("ServiceWorker.dispatchSyncEvent", r.opts, nil)
}

func init() {
	rpc.EventTypes["ServiceWorker.workerRegistrationUpdated"] = func() interface{} { return new(WorkerRegistrationUpdatedEvent) }
	rpc.EventTypes["ServiceWorker.workerVersionUpdated"] = func() interface{} { return new(WorkerVersionUpdatedEvent) }
	rpc.EventTypes["ServiceWorker.workerErrorReported"] = func() interface{} { return new(WorkerErrorReportedEvent) }
}

type WorkerRegistrationUpdatedEvent struct {
	Registrations []*ServiceWorkerRegistration `json:"registrations"`
}

type WorkerVersionUpdatedEvent struct {
	Versions []*ServiceWorkerVersion `json:"versions"`
}

type WorkerErrorReportedEvent struct {
	ErrorMessage *ServiceWorkerErrorMessage `json:"errorMessage"`
}

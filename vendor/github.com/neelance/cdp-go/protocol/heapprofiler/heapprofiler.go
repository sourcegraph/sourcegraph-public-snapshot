// (experimental)
package heapprofiler

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/runtime"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Heap snapshot object id.

type HeapSnapshotObjectId string

// Sampling Heap Profile node. Holds callsite information, allocation statistics and child nodes.

type SamplingHeapProfileNode struct {
	// Function location.
	CallFrame *runtime.CallFrame `json:"callFrame"`

	// Allocations size in bytes for the node excluding children.
	SelfSize float64 `json:"selfSize"`

	// Child nodes.
	Children []*SamplingHeapProfileNode `json:"children"`
}

// Profile.

type SamplingHeapProfile struct {
	Head *SamplingHeapProfileNode `json:"head"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("HeapProfiler.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("HeapProfiler.disable", r.opts, nil)
}

type StartTrackingHeapObjectsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) StartTrackingHeapObjects() *StartTrackingHeapObjectsRequest {
	return &StartTrackingHeapObjectsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// (optional)
func (r *StartTrackingHeapObjectsRequest) TrackAllocations(v bool) *StartTrackingHeapObjectsRequest {
	r.opts["trackAllocations"] = v
	return r
}

func (r *StartTrackingHeapObjectsRequest) Do() error {
	return r.client.Call("HeapProfiler.startTrackingHeapObjects", r.opts, nil)
}

type StopTrackingHeapObjectsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) StopTrackingHeapObjects() *StopTrackingHeapObjectsRequest {
	return &StopTrackingHeapObjectsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken when the tracking is stopped. (optional)
func (r *StopTrackingHeapObjectsRequest) ReportProgress(v bool) *StopTrackingHeapObjectsRequest {
	r.opts["reportProgress"] = v
	return r
}

func (r *StopTrackingHeapObjectsRequest) Do() error {
	return r.client.Call("HeapProfiler.stopTrackingHeapObjects", r.opts, nil)
}

type TakeHeapSnapshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) TakeHeapSnapshot() *TakeHeapSnapshotRequest {
	return &TakeHeapSnapshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken. (optional)
func (r *TakeHeapSnapshotRequest) ReportProgress(v bool) *TakeHeapSnapshotRequest {
	r.opts["reportProgress"] = v
	return r
}

func (r *TakeHeapSnapshotRequest) Do() error {
	return r.client.Call("HeapProfiler.takeHeapSnapshot", r.opts, nil)
}

type CollectGarbageRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) CollectGarbage() *CollectGarbageRequest {
	return &CollectGarbageRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *CollectGarbageRequest) Do() error {
	return r.client.Call("HeapProfiler.collectGarbage", r.opts, nil)
}

type GetObjectByHeapObjectIdRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) GetObjectByHeapObjectId() *GetObjectByHeapObjectIdRequest {
	return &GetObjectByHeapObjectIdRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetObjectByHeapObjectIdRequest) ObjectId(v HeapSnapshotObjectId) *GetObjectByHeapObjectIdRequest {
	r.opts["objectId"] = v
	return r
}

// Symbolic group name that can be used to release multiple objects. (optional)
func (r *GetObjectByHeapObjectIdRequest) ObjectGroup(v string) *GetObjectByHeapObjectIdRequest {
	r.opts["objectGroup"] = v
	return r
}

type GetObjectByHeapObjectIdResult struct {
	// Evaluation result.
	Result *runtime.RemoteObject `json:"result"`
}

func (r *GetObjectByHeapObjectIdRequest) Do() (*GetObjectByHeapObjectIdResult, error) {
	var result GetObjectByHeapObjectIdResult
	err := r.client.Call("HeapProfiler.getObjectByHeapObjectId", r.opts, &result)
	return &result, err
}

type AddInspectedHeapObjectRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables console to refer to the node with given id via $x (see Command Line API for more details $x functions).
func (d *Client) AddInspectedHeapObject() *AddInspectedHeapObjectRequest {
	return &AddInspectedHeapObjectRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Heap snapshot object id to be accessible by means of $x command line API.
func (r *AddInspectedHeapObjectRequest) HeapObjectId(v HeapSnapshotObjectId) *AddInspectedHeapObjectRequest {
	r.opts["heapObjectId"] = v
	return r
}

func (r *AddInspectedHeapObjectRequest) Do() error {
	return r.client.Call("HeapProfiler.addInspectedHeapObject", r.opts, nil)
}

type GetHeapObjectIdRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) GetHeapObjectId() *GetHeapObjectIdRequest {
	return &GetHeapObjectIdRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the object to get heap object id for.
func (r *GetHeapObjectIdRequest) ObjectId(v runtime.RemoteObjectId) *GetHeapObjectIdRequest {
	r.opts["objectId"] = v
	return r
}

type GetHeapObjectIdResult struct {
	// Id of the heap snapshot object corresponding to the passed remote object id.
	HeapSnapshotObjectId HeapSnapshotObjectId `json:"heapSnapshotObjectId"`
}

func (r *GetHeapObjectIdRequest) Do() (*GetHeapObjectIdResult, error) {
	var result GetHeapObjectIdResult
	err := r.client.Call("HeapProfiler.getHeapObjectId", r.opts, &result)
	return &result, err
}

type StartSamplingRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) StartSampling() *StartSamplingRequest {
	return &StartSamplingRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Average sample interval in bytes. Poisson distribution is used for the intervals. The default value is 32768 bytes. (optional)
func (r *StartSamplingRequest) SamplingInterval(v float64) *StartSamplingRequest {
	r.opts["samplingInterval"] = v
	return r
}

func (r *StartSamplingRequest) Do() error {
	return r.client.Call("HeapProfiler.startSampling", r.opts, nil)
}

type StopSamplingRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) StopSampling() *StopSamplingRequest {
	return &StopSamplingRequest{opts: make(map[string]interface{}), client: d.Client}
}

type StopSamplingResult struct {
	// Recorded sampling heap profile.
	Profile *SamplingHeapProfile `json:"profile"`
}

func (r *StopSamplingRequest) Do() (*StopSamplingResult, error) {
	var result StopSamplingResult
	err := r.client.Call("HeapProfiler.stopSampling", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["HeapProfiler.addHeapSnapshotChunk"] = func() interface{} { return new(AddHeapSnapshotChunkEvent) }
	rpc.EventTypes["HeapProfiler.resetProfiles"] = func() interface{} { return new(ResetProfilesEvent) }
	rpc.EventTypes["HeapProfiler.reportHeapSnapshotProgress"] = func() interface{} { return new(ReportHeapSnapshotProgressEvent) }
	rpc.EventTypes["HeapProfiler.lastSeenObjectId"] = func() interface{} { return new(LastSeenObjectIdEvent) }
	rpc.EventTypes["HeapProfiler.heapStatsUpdate"] = func() interface{} { return new(HeapStatsUpdateEvent) }
}

type AddHeapSnapshotChunkEvent struct {
	Chunk string `json:"chunk"`
}

type ResetProfilesEvent struct {
}

type ReportHeapSnapshotProgressEvent struct {
	Done int `json:"done"`

	Total int `json:"total"`

	// (optional)
	Finished bool `json:"finished"`
}

// If heap objects tracking has been started then backend regularly sends a current value for last seen object id and corresponding timestamp. If the were changes in the heap since last event then one or more heapStatsUpdate events will be sent before a new lastSeenObjectId event.
type LastSeenObjectIdEvent struct {
	LastSeenObjectId int `json:"lastSeenObjectId"`

	Timestamp float64 `json:"timestamp"`
}

// If heap objects tracking has been started then backend may send update for one or more fragments
type HeapStatsUpdateEvent struct {
	// An array of triplets. Each triplet describes a fragment. The first integer is the fragment index, the second integer is a total count of objects for the fragment, the third integer is a total size of the objects for the fragment.
	StatsUpdate []int `json:"statsUpdate"`
}

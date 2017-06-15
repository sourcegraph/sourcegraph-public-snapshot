package profiler

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/debugger"
	"github.com/neelance/cdp-go/protocol/runtime"
)

type Client struct {
	*rpc.Client
}

// Profile node. Holds callsite information, execution statistics and child nodes.

type ProfileNode struct {
	// Unique id of the node.
	Id int `json:"id"`

	// Function location.
	CallFrame *runtime.CallFrame `json:"callFrame"`

	// Number of samples where this node was on top of the call stack. (optional, experimental)
	HitCount int `json:"hitCount,omitempty"`

	// Child node ids. (optional)
	Children []int `json:"children,omitempty"`

	// The reason of being not optimized. The function may be deoptimized or marked as don't optimize. (optional)
	DeoptReason string `json:"deoptReason,omitempty"`

	// An array of source position ticks. (optional, experimental)
	PositionTicks []*PositionTickInfo `json:"positionTicks,omitempty"`
}

// Profile.

type Profile struct {
	// The list of profile nodes. First item is the root node.
	Nodes []*ProfileNode `json:"nodes"`

	// Profiling start timestamp in microseconds.
	StartTime float64 `json:"startTime"`

	// Profiling end timestamp in microseconds.
	EndTime float64 `json:"endTime"`

	// Ids of samples top nodes. (optional)
	Samples []int `json:"samples,omitempty"`

	// Time intervals between adjacent samples in microseconds. The first delta is relative to the profile startTime. (optional)
	TimeDeltas []int `json:"timeDeltas,omitempty"`
}

// Specifies a number of samples attributed to a certain source position. (experimental)

type PositionTickInfo struct {
	// Source line number (1-based).
	Line int `json:"line"`

	// Number of samples attributed to the source line.
	Ticks int `json:"ticks"`
}

// Coverage data for a source range. (experimental)

type CoverageRange struct {
	// JavaScript script source offset for the range start.
	StartOffset int `json:"startOffset"`

	// JavaScript script source offset for the range end.
	EndOffset int `json:"endOffset"`

	// Collected execution count of the source range.
	Count int `json:"count"`
}

// Coverage data for a JavaScript function. (experimental)

type FunctionCoverage struct {
	// JavaScript function name.
	FunctionName string `json:"functionName"`

	// Source ranges inside the function with coverage data.
	Ranges []*CoverageRange `json:"ranges"`
}

// Coverage data for a JavaScript script. (experimental)

type ScriptCoverage struct {
	// JavaScript script id.
	ScriptId runtime.ScriptId `json:"scriptId"`

	// JavaScript script name or url.
	URL string `json:"url"`

	// Functions contained in the script that has coverage data.
	Functions []*FunctionCoverage `json:"functions"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Profiler.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Profiler.disable", r.opts, nil)
}

type SetSamplingIntervalRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Changes CPU profiler sampling interval. Must be called before CPU profiles recording started.
func (d *Client) SetSamplingInterval() *SetSamplingIntervalRequest {
	return &SetSamplingIntervalRequest{opts: make(map[string]interface{}), client: d.Client}
}

// New sampling interval in microseconds.
func (r *SetSamplingIntervalRequest) Interval(v int) *SetSamplingIntervalRequest {
	r.opts["interval"] = v
	return r
}

func (r *SetSamplingIntervalRequest) Do() error {
	return r.client.Call("Profiler.setSamplingInterval", r.opts, nil)
}

type StartRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Start() *StartRequest {
	return &StartRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *StartRequest) Do() error {
	return r.client.Call("Profiler.start", r.opts, nil)
}

type StopRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Stop() *StopRequest {
	return &StopRequest{opts: make(map[string]interface{}), client: d.Client}
}

type StopResult struct {
	// Recorded profile.
	Profile *Profile `json:"profile"`
}

func (r *StopRequest) Do() (*StopResult, error) {
	var result StopResult
	err := r.client.Call("Profiler.stop", r.opts, &result)
	return &result, err
}

type StartPreciseCoverageRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enable precise code coverage. Coverage data for JavaScript executed before enabling precise code coverage may be incomplete. Enabling prevents running optimized code and resets execution counters. (experimental)
func (d *Client) StartPreciseCoverage() *StartPreciseCoverageRequest {
	return &StartPreciseCoverageRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Collect accurate call counts beyond simple 'covered' or 'not covered'. (optional)
func (r *StartPreciseCoverageRequest) CallCount(v bool) *StartPreciseCoverageRequest {
	r.opts["callCount"] = v
	return r
}

func (r *StartPreciseCoverageRequest) Do() error {
	return r.client.Call("Profiler.startPreciseCoverage", r.opts, nil)
}

type StopPreciseCoverageRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disable precise code coverage. Disabling releases unnecessary execution count records and allows executing optimized code. (experimental)
func (d *Client) StopPreciseCoverage() *StopPreciseCoverageRequest {
	return &StopPreciseCoverageRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *StopPreciseCoverageRequest) Do() error {
	return r.client.Call("Profiler.stopPreciseCoverage", r.opts, nil)
}

type TakePreciseCoverageRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Collect coverage data for the current isolate, and resets execution counters. Precise code coverage needs to have started. (experimental)
func (d *Client) TakePreciseCoverage() *TakePreciseCoverageRequest {
	return &TakePreciseCoverageRequest{opts: make(map[string]interface{}), client: d.Client}
}

type TakePreciseCoverageResult struct {
	// Coverage data for the current isolate.
	Result []*ScriptCoverage `json:"result"`
}

func (r *TakePreciseCoverageRequest) Do() (*TakePreciseCoverageResult, error) {
	var result TakePreciseCoverageResult
	err := r.client.Call("Profiler.takePreciseCoverage", r.opts, &result)
	return &result, err
}

type GetBestEffortCoverageRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Collect coverage data for the current isolate. The coverage data may be incomplete due to garbage collection. (experimental)
func (d *Client) GetBestEffortCoverage() *GetBestEffortCoverageRequest {
	return &GetBestEffortCoverageRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetBestEffortCoverageResult struct {
	// Coverage data for the current isolate.
	Result []*ScriptCoverage `json:"result"`
}

func (r *GetBestEffortCoverageRequest) Do() (*GetBestEffortCoverageResult, error) {
	var result GetBestEffortCoverageResult
	err := r.client.Call("Profiler.getBestEffortCoverage", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["Profiler.consoleProfileStarted"] = func() interface{} { return new(ConsoleProfileStartedEvent) }
	rpc.EventTypes["Profiler.consoleProfileFinished"] = func() interface{} { return new(ConsoleProfileFinishedEvent) }
}

// Sent when new profile recording is started using console.profile() call.
type ConsoleProfileStartedEvent struct {
	Id string `json:"id"`

	// Location of console.profile().
	Location *debugger.Location `json:"location"`

	// Profile title passed as an argument to console.profile(). (optional)
	Title string `json:"title"`
}

type ConsoleProfileFinishedEvent struct {
	Id string `json:"id"`

	// Location of console.profileEnd().
	Location *debugger.Location `json:"location"`

	Profile *Profile `json:"profile"`

	// Profile title passed as an argument to console.profile(). (optional)
	Title string `json:"title"`
}

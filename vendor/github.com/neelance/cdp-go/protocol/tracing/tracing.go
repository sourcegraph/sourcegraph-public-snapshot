// (experimental)
package tracing

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/io"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Configuration for memory dump. Used only when "memory-infra" category is enabled.

type MemoryDumpConfig struct {
}

type TraceConfig struct {
	// Controls how the trace buffer stores data. (optional)
	RecordMode string `json:"recordMode,omitempty"`

	// Turns on JavaScript stack sampling. (optional)
	EnableSampling bool `json:"enableSampling,omitempty"`

	// Turns on system tracing. (optional)
	EnableSystrace bool `json:"enableSystrace,omitempty"`

	// Turns on argument filter. (optional)
	EnableArgumentFilter bool `json:"enableArgumentFilter,omitempty"`

	// Included category filters. (optional)
	IncludedCategories []string `json:"includedCategories,omitempty"`

	// Excluded category filters. (optional)
	ExcludedCategories []string `json:"excludedCategories,omitempty"`

	// Configuration to synthesize the delays in tracing. (optional)
	SyntheticDelays []string `json:"syntheticDelays,omitempty"`

	// Configuration for memory dump triggers. Used only when "memory-infra" category is enabled. (optional)
	MemoryDumpConfig *MemoryDumpConfig `json:"memoryDumpConfig,omitempty"`
}

type StartRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Start trace events collection.
func (d *Client) Start() *StartRequest {
	return &StartRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Category/tag filter (optional)
func (r *StartRequest) Categories(v string) *StartRequest {
	r.opts["categories"] = v
	return r
}

// Tracing options (optional)
func (r *StartRequest) Options(v string) *StartRequest {
	r.opts["options"] = v
	return r
}

// If set, the agent will issue bufferUsage events at this interval, specified in milliseconds (optional)
func (r *StartRequest) BufferUsageReportingInterval(v float64) *StartRequest {
	r.opts["bufferUsageReportingInterval"] = v
	return r
}

// Whether to report trace events as series of dataCollected events or to save trace to a stream (defaults to <code>ReportEvents</code>). (optional)
func (r *StartRequest) TransferMode(v string) *StartRequest {
	r.opts["transferMode"] = v
	return r
}

// (optional)
func (r *StartRequest) TraceConfig(v *TraceConfig) *StartRequest {
	r.opts["traceConfig"] = v
	return r
}

func (r *StartRequest) Do() error {
	return r.client.Call("Tracing.start", r.opts, nil)
}

type EndRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Stop trace events collection.
func (d *Client) End() *EndRequest {
	return &EndRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EndRequest) Do() error {
	return r.client.Call("Tracing.end", r.opts, nil)
}

type GetCategoriesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Gets supported tracing categories.
func (d *Client) GetCategories() *GetCategoriesRequest {
	return &GetCategoriesRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetCategoriesResult struct {
	// A list of supported tracing categories.
	Categories []string `json:"categories"`
}

func (r *GetCategoriesRequest) Do() (*GetCategoriesResult, error) {
	var result GetCategoriesResult
	err := r.client.Call("Tracing.getCategories", r.opts, &result)
	return &result, err
}

type RequestMemoryDumpRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Request a global memory dump.
func (d *Client) RequestMemoryDump() *RequestMemoryDumpRequest {
	return &RequestMemoryDumpRequest{opts: make(map[string]interface{}), client: d.Client}
}

type RequestMemoryDumpResult struct {
	// GUID of the resulting global memory dump.
	DumpGuid string `json:"dumpGuid"`

	// True iff the global memory dump succeeded.
	Success bool `json:"success"`
}

func (r *RequestMemoryDumpRequest) Do() (*RequestMemoryDumpResult, error) {
	var result RequestMemoryDumpResult
	err := r.client.Call("Tracing.requestMemoryDump", r.opts, &result)
	return &result, err
}

type RecordClockSyncMarkerRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Record a clock sync marker in the trace.
func (d *Client) RecordClockSyncMarker() *RecordClockSyncMarkerRequest {
	return &RecordClockSyncMarkerRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The ID of this clock sync marker
func (r *RecordClockSyncMarkerRequest) SyncId(v string) *RecordClockSyncMarkerRequest {
	r.opts["syncId"] = v
	return r
}

func (r *RecordClockSyncMarkerRequest) Do() error {
	return r.client.Call("Tracing.recordClockSyncMarker", r.opts, nil)
}

func init() {
	rpc.EventTypes["Tracing.dataCollected"] = func() interface{} { return new(DataCollectedEvent) }
	rpc.EventTypes["Tracing.tracingComplete"] = func() interface{} { return new(TracingCompleteEvent) }
	rpc.EventTypes["Tracing.bufferUsage"] = func() interface{} { return new(BufferUsageEvent) }
}

// Contains an bucket of collected trace events. When tracing is stopped collected events will be send as a sequence of dataCollected events followed by tracingComplete event.
type DataCollectedEvent struct {
	Value []interface{} `json:"value"`
}

// Signals that tracing is stopped and there is no trace buffers pending flush, all data were delivered via dataCollected events.
type TracingCompleteEvent struct {
	// A handle of the stream that holds resulting trace data. (optional)
	Stream io.StreamHandle `json:"stream"`
}

type BufferUsageEvent struct {
	// A number in range [0..1] that indicates the used size of event buffer as a fraction of its total size. (optional)
	PercentFull float64 `json:"percentFull"`

	// An approximate number of events in the trace log. (optional)
	EventCount float64 `json:"eventCount"`

	// A number in range [0..1] that indicates the used size of event buffer as a fraction of its total size. (optional)
	Value float64 `json:"value"`
}

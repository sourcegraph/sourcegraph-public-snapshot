// (experimental)
package layertree

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/dom"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Unique Layer identifier.

type LayerId string

// Unique snapshot identifier.

type SnapshotId string

// Rectangle where scrolling happens on the main thread.

type ScrollRect struct {
	// Rectangle itself.
	Rect *dom.Rect `json:"rect"`

	// Reason for rectangle to force scrolling on the main thread
	Type string `json:"type"`
}

// Serialized fragment of layer picture along with its offset within the layer.

type PictureTile struct {
	// Offset from owning layer left boundary
	X float64 `json:"x"`

	// Offset from owning layer top boundary
	Y float64 `json:"y"`

	// Base64-encoded snapshot data.
	Picture string `json:"picture"`
}

// Information about a compositing layer.

type Layer struct {
	// The unique id for this layer.
	LayerId LayerId `json:"layerId"`

	// The id of parent (not present for root). (optional)
	ParentLayerId LayerId `json:"parentLayerId,omitempty"`

	// The backend id for the node associated with this layer. (optional)
	BackendNodeId dom.BackendNodeId `json:"backendNodeId,omitempty"`

	// Offset from parent layer, X coordinate.
	OffsetX float64 `json:"offsetX"`

	// Offset from parent layer, Y coordinate.
	OffsetY float64 `json:"offsetY"`

	// Layer width.
	Width float64 `json:"width"`

	// Layer height.
	Height float64 `json:"height"`

	// Transformation matrix for layer, default is identity matrix (optional)
	Transform []float64 `json:"transform,omitempty"`

	// Transform anchor point X, absent if no transform specified (optional)
	AnchorX float64 `json:"anchorX,omitempty"`

	// Transform anchor point Y, absent if no transform specified (optional)
	AnchorY float64 `json:"anchorY,omitempty"`

	// Transform anchor point Z, absent if no transform specified (optional)
	AnchorZ float64 `json:"anchorZ,omitempty"`

	// Indicates how many time this layer has painted.
	PaintCount int `json:"paintCount"`

	// Indicates whether this layer hosts any content, rather than being used for transform/scrolling purposes only.
	DrawsContent bool `json:"drawsContent"`

	// Set if layer is not visible. (optional)
	Invisible bool `json:"invisible,omitempty"`

	// Rectangles scrolling on main thread only. (optional)
	ScrollRects []*ScrollRect `json:"scrollRects,omitempty"`
}

// Array of timings, one per paint step.

type PaintProfile []float64

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables compositing tree inspection.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("LayerTree.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables compositing tree inspection.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("LayerTree.disable", r.opts, nil)
}

type CompositingReasonsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Provides the reasons why the given layer was composited.
func (d *Client) CompositingReasons() *CompositingReasonsRequest {
	return &CompositingReasonsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The id of the layer for which we want to get the reasons it was composited.
func (r *CompositingReasonsRequest) LayerId(v LayerId) *CompositingReasonsRequest {
	r.opts["layerId"] = v
	return r
}

type CompositingReasonsResult struct {
	// A list of strings specifying reasons for the given layer to become composited.
	CompositingReasons []string `json:"compositingReasons"`
}

func (r *CompositingReasonsRequest) Do() (*CompositingReasonsResult, error) {
	var result CompositingReasonsResult
	err := r.client.Call("LayerTree.compositingReasons", r.opts, &result)
	return &result, err
}

type MakeSnapshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns the layer snapshot identifier.
func (d *Client) MakeSnapshot() *MakeSnapshotRequest {
	return &MakeSnapshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The id of the layer.
func (r *MakeSnapshotRequest) LayerId(v LayerId) *MakeSnapshotRequest {
	r.opts["layerId"] = v
	return r
}

type MakeSnapshotResult struct {
	// The id of the layer snapshot.
	SnapshotId SnapshotId `json:"snapshotId"`
}

func (r *MakeSnapshotRequest) Do() (*MakeSnapshotResult, error) {
	var result MakeSnapshotResult
	err := r.client.Call("LayerTree.makeSnapshot", r.opts, &result)
	return &result, err
}

type LoadSnapshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns the snapshot identifier.
func (d *Client) LoadSnapshot() *LoadSnapshotRequest {
	return &LoadSnapshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// An array of tiles composing the snapshot.
func (r *LoadSnapshotRequest) Tiles(v []*PictureTile) *LoadSnapshotRequest {
	r.opts["tiles"] = v
	return r
}

type LoadSnapshotResult struct {
	// The id of the snapshot.
	SnapshotId SnapshotId `json:"snapshotId"`
}

func (r *LoadSnapshotRequest) Do() (*LoadSnapshotResult, error) {
	var result LoadSnapshotResult
	err := r.client.Call("LayerTree.loadSnapshot", r.opts, &result)
	return &result, err
}

type ReleaseSnapshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Releases layer snapshot captured by the back-end.
func (d *Client) ReleaseSnapshot() *ReleaseSnapshotRequest {
	return &ReleaseSnapshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The id of the layer snapshot.
func (r *ReleaseSnapshotRequest) SnapshotId(v SnapshotId) *ReleaseSnapshotRequest {
	r.opts["snapshotId"] = v
	return r
}

func (r *ReleaseSnapshotRequest) Do() error {
	return r.client.Call("LayerTree.releaseSnapshot", r.opts, nil)
}

type ProfileSnapshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) ProfileSnapshot() *ProfileSnapshotRequest {
	return &ProfileSnapshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The id of the layer snapshot.
func (r *ProfileSnapshotRequest) SnapshotId(v SnapshotId) *ProfileSnapshotRequest {
	r.opts["snapshotId"] = v
	return r
}

// The maximum number of times to replay the snapshot (1, if not specified). (optional)
func (r *ProfileSnapshotRequest) MinRepeatCount(v int) *ProfileSnapshotRequest {
	r.opts["minRepeatCount"] = v
	return r
}

// The minimum duration (in seconds) to replay the snapshot. (optional)
func (r *ProfileSnapshotRequest) MinDuration(v float64) *ProfileSnapshotRequest {
	r.opts["minDuration"] = v
	return r
}

// The clip rectangle to apply when replaying the snapshot. (optional)
func (r *ProfileSnapshotRequest) ClipRect(v *dom.Rect) *ProfileSnapshotRequest {
	r.opts["clipRect"] = v
	return r
}

type ProfileSnapshotResult struct {
	// The array of paint profiles, one per run.
	Timings []PaintProfile `json:"timings"`
}

func (r *ProfileSnapshotRequest) Do() (*ProfileSnapshotResult, error) {
	var result ProfileSnapshotResult
	err := r.client.Call("LayerTree.profileSnapshot", r.opts, &result)
	return &result, err
}

type ReplaySnapshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Replays the layer snapshot and returns the resulting bitmap.
func (d *Client) ReplaySnapshot() *ReplaySnapshotRequest {
	return &ReplaySnapshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The id of the layer snapshot.
func (r *ReplaySnapshotRequest) SnapshotId(v SnapshotId) *ReplaySnapshotRequest {
	r.opts["snapshotId"] = v
	return r
}

// The first step to replay from (replay from the very start if not specified). (optional)
func (r *ReplaySnapshotRequest) FromStep(v int) *ReplaySnapshotRequest {
	r.opts["fromStep"] = v
	return r
}

// The last step to replay to (replay till the end if not specified). (optional)
func (r *ReplaySnapshotRequest) ToStep(v int) *ReplaySnapshotRequest {
	r.opts["toStep"] = v
	return r
}

// The scale to apply while replaying (defaults to 1). (optional)
func (r *ReplaySnapshotRequest) Scale(v float64) *ReplaySnapshotRequest {
	r.opts["scale"] = v
	return r
}

type ReplaySnapshotResult struct {
	// A data: URL for resulting image.
	DataURL string `json:"dataURL"`
}

func (r *ReplaySnapshotRequest) Do() (*ReplaySnapshotResult, error) {
	var result ReplaySnapshotResult
	err := r.client.Call("LayerTree.replaySnapshot", r.opts, &result)
	return &result, err
}

type SnapshotCommandLogRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Replays the layer snapshot and returns canvas log.
func (d *Client) SnapshotCommandLog() *SnapshotCommandLogRequest {
	return &SnapshotCommandLogRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The id of the layer snapshot.
func (r *SnapshotCommandLogRequest) SnapshotId(v SnapshotId) *SnapshotCommandLogRequest {
	r.opts["snapshotId"] = v
	return r
}

type SnapshotCommandLogResult struct {
	// The array of canvas function calls.
	CommandLog []interface{} `json:"commandLog"`
}

func (r *SnapshotCommandLogRequest) Do() (*SnapshotCommandLogResult, error) {
	var result SnapshotCommandLogResult
	err := r.client.Call("LayerTree.snapshotCommandLog", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["LayerTree.layerTreeDidChange"] = func() interface{} { return new(LayerTreeDidChangeEvent) }
	rpc.EventTypes["LayerTree.layerPainted"] = func() interface{} { return new(LayerPaintedEvent) }
}

type LayerTreeDidChangeEvent struct {
	// Layer tree, absent if not in the comspositing mode. (optional)
	Layers []*Layer `json:"layers"`
}

type LayerPaintedEvent struct {
	// The id of the painted layer.
	LayerId LayerId `json:"layerId"`

	// Clip rectangle.
	Clip *dom.Rect `json:"clip"`
}

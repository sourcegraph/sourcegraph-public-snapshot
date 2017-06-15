// (experimental)
package animation

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/dom"
	"github.com/neelance/cdp-go/protocol/runtime"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Animation instance. (experimental)

type Animation struct {
	// <code>Animation</code>'s id.
	Id string `json:"id"`

	// <code>Animation</code>'s name.
	Name string `json:"name"`

	// <code>Animation</code>'s internal paused state.
	PausedState bool `json:"pausedState"`

	// <code>Animation</code>'s play state.
	PlayState string `json:"playState"`

	// <code>Animation</code>'s playback rate.
	PlaybackRate float64 `json:"playbackRate"`

	// <code>Animation</code>'s start time.
	StartTime float64 `json:"startTime"`

	// <code>Animation</code>'s current time.
	CurrentTime float64 `json:"currentTime"`

	// <code>Animation</code>'s source animation node.
	Source *AnimationEffect `json:"source"`

	// Animation type of <code>Animation</code>.
	Type string `json:"type"`

	// A unique ID for <code>Animation</code> representing the sources that triggered this CSS animation/transition. (optional)
	CssId string `json:"cssId,omitempty"`
}

// AnimationEffect instance (experimental)

type AnimationEffect struct {
	// <code>AnimationEffect</code>'s delay.
	Delay float64 `json:"delay"`

	// <code>AnimationEffect</code>'s end delay.
	EndDelay float64 `json:"endDelay"`

	// <code>AnimationEffect</code>'s iteration start.
	IterationStart float64 `json:"iterationStart"`

	// <code>AnimationEffect</code>'s iterations.
	Iterations float64 `json:"iterations"`

	// <code>AnimationEffect</code>'s iteration duration.
	Duration float64 `json:"duration"`

	// <code>AnimationEffect</code>'s playback direction.
	Direction string `json:"direction"`

	// <code>AnimationEffect</code>'s fill mode.
	Fill string `json:"fill"`

	// <code>AnimationEffect</code>'s target node.
	BackendNodeId dom.BackendNodeId `json:"backendNodeId"`

	// <code>AnimationEffect</code>'s keyframes. (optional)
	KeyframesRule *KeyframesRule `json:"keyframesRule,omitempty"`

	// <code>AnimationEffect</code>'s timing function.
	Easing string `json:"easing"`
}

// Keyframes Rule

type KeyframesRule struct {
	// CSS keyframed animation's name. (optional)
	Name string `json:"name,omitempty"`

	// List of animation keyframes.
	Keyframes []*KeyframeStyle `json:"keyframes"`
}

// Keyframe Style

type KeyframeStyle struct {
	// Keyframe's time offset.
	Offset string `json:"offset"`

	// <code>AnimationEffect</code>'s timing function.
	Easing string `json:"easing"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables animation domain notifications.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Animation.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables animation domain notifications.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Animation.disable", r.opts, nil)
}

type GetPlaybackRateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Gets the playback rate of the document timeline.
func (d *Client) GetPlaybackRate() *GetPlaybackRateRequest {
	return &GetPlaybackRateRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetPlaybackRateResult struct {
	// Playback rate for animations on page.
	PlaybackRate float64 `json:"playbackRate"`
}

func (r *GetPlaybackRateRequest) Do() (*GetPlaybackRateResult, error) {
	var result GetPlaybackRateResult
	err := r.client.Call("Animation.getPlaybackRate", r.opts, &result)
	return &result, err
}

type SetPlaybackRateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets the playback rate of the document timeline.
func (d *Client) SetPlaybackRate() *SetPlaybackRateRequest {
	return &SetPlaybackRateRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Playback rate for animations on page
func (r *SetPlaybackRateRequest) PlaybackRate(v float64) *SetPlaybackRateRequest {
	r.opts["playbackRate"] = v
	return r
}

func (r *SetPlaybackRateRequest) Do() error {
	return r.client.Call("Animation.setPlaybackRate", r.opts, nil)
}

type GetCurrentTimeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns the current time of the an animation.
func (d *Client) GetCurrentTime() *GetCurrentTimeRequest {
	return &GetCurrentTimeRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Id of animation.
func (r *GetCurrentTimeRequest) Id(v string) *GetCurrentTimeRequest {
	r.opts["id"] = v
	return r
}

type GetCurrentTimeResult struct {
	// Current time of the page.
	CurrentTime float64 `json:"currentTime"`
}

func (r *GetCurrentTimeRequest) Do() (*GetCurrentTimeResult, error) {
	var result GetCurrentTimeResult
	err := r.client.Call("Animation.getCurrentTime", r.opts, &result)
	return &result, err
}

type SetPausedRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets the paused state of a set of animations.
func (d *Client) SetPaused() *SetPausedRequest {
	return &SetPausedRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Animations to set the pause state of.
func (r *SetPausedRequest) Animations(v []string) *SetPausedRequest {
	r.opts["animations"] = v
	return r
}

// Paused state to set to.
func (r *SetPausedRequest) Paused(v bool) *SetPausedRequest {
	r.opts["paused"] = v
	return r
}

func (r *SetPausedRequest) Do() error {
	return r.client.Call("Animation.setPaused", r.opts, nil)
}

type SetTimingRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets the timing of an animation node.
func (d *Client) SetTiming() *SetTimingRequest {
	return &SetTimingRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Animation id.
func (r *SetTimingRequest) AnimationId(v string) *SetTimingRequest {
	r.opts["animationId"] = v
	return r
}

// Duration of the animation.
func (r *SetTimingRequest) Duration(v float64) *SetTimingRequest {
	r.opts["duration"] = v
	return r
}

// Delay of the animation.
func (r *SetTimingRequest) Delay(v float64) *SetTimingRequest {
	r.opts["delay"] = v
	return r
}

func (r *SetTimingRequest) Do() error {
	return r.client.Call("Animation.setTiming", r.opts, nil)
}

type SeekAnimationsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Seek a set of animations to a particular time within each animation.
func (d *Client) SeekAnimations() *SeekAnimationsRequest {
	return &SeekAnimationsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// List of animation ids to seek.
func (r *SeekAnimationsRequest) Animations(v []string) *SeekAnimationsRequest {
	r.opts["animations"] = v
	return r
}

// Set the current time of each animation.
func (r *SeekAnimationsRequest) CurrentTime(v float64) *SeekAnimationsRequest {
	r.opts["currentTime"] = v
	return r
}

func (r *SeekAnimationsRequest) Do() error {
	return r.client.Call("Animation.seekAnimations", r.opts, nil)
}

type ReleaseAnimationsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Releases a set of animations to no longer be manipulated.
func (d *Client) ReleaseAnimations() *ReleaseAnimationsRequest {
	return &ReleaseAnimationsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// List of animation ids to seek.
func (r *ReleaseAnimationsRequest) Animations(v []string) *ReleaseAnimationsRequest {
	r.opts["animations"] = v
	return r
}

func (r *ReleaseAnimationsRequest) Do() error {
	return r.client.Call("Animation.releaseAnimations", r.opts, nil)
}

type ResolveAnimationRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Gets the remote object of the Animation.
func (d *Client) ResolveAnimation() *ResolveAnimationRequest {
	return &ResolveAnimationRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Animation id.
func (r *ResolveAnimationRequest) AnimationId(v string) *ResolveAnimationRequest {
	r.opts["animationId"] = v
	return r
}

type ResolveAnimationResult struct {
	// Corresponding remote object.
	RemoteObject *runtime.RemoteObject `json:"remoteObject"`
}

func (r *ResolveAnimationRequest) Do() (*ResolveAnimationResult, error) {
	var result ResolveAnimationResult
	err := r.client.Call("Animation.resolveAnimation", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["Animation.animationCreated"] = func() interface{} { return new(AnimationCreatedEvent) }
	rpc.EventTypes["Animation.animationStarted"] = func() interface{} { return new(AnimationStartedEvent) }
	rpc.EventTypes["Animation.animationCanceled"] = func() interface{} { return new(AnimationCanceledEvent) }
}

// Event for each animation that has been created.
type AnimationCreatedEvent struct {
	// Id of the animation that was created.
	Id string `json:"id"`
}

// Event for animation that has been started.
type AnimationStartedEvent struct {
	// Animation that was started.
	Animation *Animation `json:"animation"`
}

// Event for when an animation has been cancelled.
type AnimationCanceledEvent struct {
	// Id of the animation that was cancelled.
	Id string `json:"id"`
}

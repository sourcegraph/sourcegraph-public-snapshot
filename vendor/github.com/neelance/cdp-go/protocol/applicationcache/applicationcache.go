// (experimental)
package applicationcache

import (
	"github.com/neelance/cdp-go/rpc"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Detailed application cache resource information.

type ApplicationCacheResource struct {
	// Resource url.
	URL string `json:"url"`

	// Resource size.
	Size int `json:"size"`

	// Resource type.
	Type string `json:"type"`
}

// Detailed application cache information.

type ApplicationCache struct {
	// Manifest URL.
	ManifestURL string `json:"manifestURL"`

	// Application cache size.
	Size float64 `json:"size"`

	// Application cache creation time.
	CreationTime float64 `json:"creationTime"`

	// Application cache update time.
	UpdateTime float64 `json:"updateTime"`

	// Application cache resources.
	Resources []*ApplicationCacheResource `json:"resources"`
}

// Frame identifier - manifest URL pair.

type FrameWithManifest struct {
	// Frame identifier.
	FrameId string `json:"frameId"`

	// Manifest URL.
	ManifestURL string `json:"manifestURL"`

	// Application cache status.
	Status int `json:"status"`
}

type GetFramesWithManifestsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns array of frame identifiers with manifest urls for each frame containing a document associated with some application cache.
func (d *Client) GetFramesWithManifests() *GetFramesWithManifestsRequest {
	return &GetFramesWithManifestsRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetFramesWithManifestsResult struct {
	// Array of frame identifiers with manifest urls for each frame containing a document associated with some application cache.
	FrameIds []*FrameWithManifest `json:"frameIds"`
}

func (r *GetFramesWithManifestsRequest) Do() (*GetFramesWithManifestsResult, error) {
	var result GetFramesWithManifestsResult
	err := r.client.Call("ApplicationCache.getFramesWithManifests", r.opts, &result)
	return &result, err
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables application cache domain notifications.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("ApplicationCache.enable", r.opts, nil)
}

type GetManifestForFrameRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns manifest URL for document in the given frame.
func (d *Client) GetManifestForFrame() *GetManifestForFrameRequest {
	return &GetManifestForFrameRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the frame containing document whose manifest is retrieved.
func (r *GetManifestForFrameRequest) FrameId(v string) *GetManifestForFrameRequest {
	r.opts["frameId"] = v
	return r
}

type GetManifestForFrameResult struct {
	// Manifest URL for document in the given frame.
	ManifestURL string `json:"manifestURL"`
}

func (r *GetManifestForFrameRequest) Do() (*GetManifestForFrameResult, error) {
	var result GetManifestForFrameResult
	err := r.client.Call("ApplicationCache.getManifestForFrame", r.opts, &result)
	return &result, err
}

type GetApplicationCacheForFrameRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns relevant application cache data for the document in given frame.
func (d *Client) GetApplicationCacheForFrame() *GetApplicationCacheForFrameRequest {
	return &GetApplicationCacheForFrameRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the frame containing document whose application cache is retrieved.
func (r *GetApplicationCacheForFrameRequest) FrameId(v string) *GetApplicationCacheForFrameRequest {
	r.opts["frameId"] = v
	return r
}

type GetApplicationCacheForFrameResult struct {
	// Relevant application cache data for the document in given frame.
	ApplicationCache *ApplicationCache `json:"applicationCache"`
}

func (r *GetApplicationCacheForFrameRequest) Do() (*GetApplicationCacheForFrameResult, error) {
	var result GetApplicationCacheForFrameResult
	err := r.client.Call("ApplicationCache.getApplicationCacheForFrame", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["ApplicationCache.applicationCacheStatusUpdated"] = func() interface{} { return new(ApplicationCacheStatusUpdatedEvent) }
	rpc.EventTypes["ApplicationCache.networkStateUpdated"] = func() interface{} { return new(NetworkStateUpdatedEvent) }
}

type ApplicationCacheStatusUpdatedEvent struct {
	// Identifier of the frame containing document whose application cache updated status.
	FrameId string `json:"frameId"`

	// Manifest URL.
	ManifestURL string `json:"manifestURL"`

	// Updated application cache status.
	Status int `json:"status"`
}

type NetworkStateUpdatedEvent struct {
	IsNowOnline bool `json:"isNowOnline"`
}

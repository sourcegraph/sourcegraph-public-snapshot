// Actions and events related to the inspected page belong to the page domain.
package page

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/debugger"
	"github.com/neelance/cdp-go/protocol/dom"
	"github.com/neelance/cdp-go/protocol/emulation"
	"github.com/neelance/cdp-go/protocol/network"
	"github.com/neelance/cdp-go/protocol/runtime"
)

// Actions and events related to the inspected page belong to the page domain.
type Client struct {
	*rpc.Client
}

// Resource type as it was perceived by the rendering engine.

type ResourceType string

// Unique frame identifier.

type FrameId string

// Information about the Frame on the page.

type Frame struct {
	// Frame unique identifier.
	Id string `json:"id"`

	// Parent frame identifier. (optional)
	ParentId string `json:"parentId,omitempty"`

	// Identifier of the loader associated with this frame.
	LoaderId network.LoaderId `json:"loaderId"`

	// Frame's name as specified in the tag. (optional)
	Name string `json:"name,omitempty"`

	// Frame document's URL.
	URL string `json:"url"`

	// Frame document's security origin.
	SecurityOrigin string `json:"securityOrigin"`

	// Frame document's mimeType as determined by the browser.
	MimeType string `json:"mimeType"`
}

// Information about the Resource on the page. (experimental)

type FrameResource struct {
	// Resource URL.
	URL string `json:"url"`

	// Type of this resource.
	Type ResourceType `json:"type"`

	// Resource mimeType as determined by the browser.
	MimeType string `json:"mimeType"`

	// last-modified timestamp as reported by server. (optional)
	LastModified network.Timestamp `json:"lastModified,omitempty"`

	// Resource content size. (optional)
	ContentSize float64 `json:"contentSize,omitempty"`

	// True if the resource failed to load. (optional)
	Failed bool `json:"failed,omitempty"`

	// True if the resource was canceled during loading. (optional)
	Canceled bool `json:"canceled,omitempty"`
}

// Information about the Frame hierarchy along with their cached resources. (experimental)

type FrameResourceTree struct {
	// Frame information for this tree item.
	Frame *Frame `json:"frame"`

	// Child frames. (optional)
	ChildFrames []*FrameResourceTree `json:"childFrames,omitempty"`

	// Information about frame resources.
	Resources []*FrameResource `json:"resources"`
}

// Unique script identifier. (experimental)

type ScriptIdentifier string

// Transition type. (experimental)

type TransitionType string

// Navigation history entry. (experimental)

type NavigationEntry struct {
	// Unique id of the navigation history entry.
	Id int `json:"id"`

	// URL of the navigation history entry.
	URL string `json:"url"`

	// URL that the user typed in the url bar.
	UserTypedURL string `json:"userTypedURL"`

	// Title of the navigation history entry.
	Title string `json:"title"`

	// Transition type.
	TransitionType TransitionType `json:"transitionType"`
}

// Screencast frame metadata. (experimental)

type ScreencastFrameMetadata struct {
	// Top offset in DIP.
	OffsetTop float64 `json:"offsetTop"`

	// Page scale factor.
	PageScaleFactor float64 `json:"pageScaleFactor"`

	// Device screen width in DIP.
	DeviceWidth float64 `json:"deviceWidth"`

	// Device screen height in DIP.
	DeviceHeight float64 `json:"deviceHeight"`

	// Position of horizontal scroll in CSS pixels.
	ScrollOffsetX float64 `json:"scrollOffsetX"`

	// Position of vertical scroll in CSS pixels.
	ScrollOffsetY float64 `json:"scrollOffsetY"`

	// Frame swap timestamp. (optional, experimental)
	Timestamp float64 `json:"timestamp,omitempty"`
}

// Javascript dialog type. (experimental)

type DialogType string

// Error while paring app manifest. (experimental)

type AppManifestError struct {
	// Error message.
	Message string `json:"message"`

	// If criticial, this is a non-recoverable parse error.
	Critical int `json:"critical"`

	// Error line.
	Line int `json:"line"`

	// Error column.
	Column int `json:"column"`
}

// Proceed: allow the navigation; Cancel: cancel the navigation; CancelAndIgnore: cancels the navigation and makes the requester of the navigation acts like the request was never made. (experimental)

type NavigationResponse string

// Layout viewport position and dimensions. (experimental)

type LayoutViewport struct {
	// Horizontal offset relative to the document (CSS pixels).
	PageX int `json:"pageX"`

	// Vertical offset relative to the document (CSS pixels).
	PageY int `json:"pageY"`

	// Width (CSS pixels), excludes scrollbar if present.
	ClientWidth int `json:"clientWidth"`

	// Height (CSS pixels), excludes scrollbar if present.
	ClientHeight int `json:"clientHeight"`
}

// Visual viewport position, dimensions, and scale. (experimental)

type VisualViewport struct {
	// Horizontal offset relative to the layout viewport (CSS pixels).
	OffsetX float64 `json:"offsetX"`

	// Vertical offset relative to the layout viewport (CSS pixels).
	OffsetY float64 `json:"offsetY"`

	// Horizontal offset relative to the document (CSS pixels).
	PageX float64 `json:"pageX"`

	// Vertical offset relative to the document (CSS pixels).
	PageY float64 `json:"pageY"`

	// Width (CSS pixels), excludes scrollbar if present.
	ClientWidth float64 `json:"clientWidth"`

	// Height (CSS pixels), excludes scrollbar if present.
	ClientHeight float64 `json:"clientHeight"`

	// Scale relative to the ideal viewport (size at width=device-width).
	Scale float64 `json:"scale"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables page domain notifications.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Page.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables page domain notifications.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Page.disable", r.opts, nil)
}

type AddScriptToEvaluateOnLoadRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// (experimental)
func (d *Client) AddScriptToEvaluateOnLoad() *AddScriptToEvaluateOnLoadRequest {
	return &AddScriptToEvaluateOnLoadRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *AddScriptToEvaluateOnLoadRequest) ScriptSource(v string) *AddScriptToEvaluateOnLoadRequest {
	r.opts["scriptSource"] = v
	return r
}

type AddScriptToEvaluateOnLoadResult struct {
	// Identifier of the added script.
	Identifier ScriptIdentifier `json:"identifier"`
}

func (r *AddScriptToEvaluateOnLoadRequest) Do() (*AddScriptToEvaluateOnLoadResult, error) {
	var result AddScriptToEvaluateOnLoadResult
	err := r.client.Call("Page.addScriptToEvaluateOnLoad", r.opts, &result)
	return &result, err
}

type RemoveScriptToEvaluateOnLoadRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// (experimental)
func (d *Client) RemoveScriptToEvaluateOnLoad() *RemoveScriptToEvaluateOnLoadRequest {
	return &RemoveScriptToEvaluateOnLoadRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *RemoveScriptToEvaluateOnLoadRequest) Identifier(v ScriptIdentifier) *RemoveScriptToEvaluateOnLoadRequest {
	r.opts["identifier"] = v
	return r
}

func (r *RemoveScriptToEvaluateOnLoadRequest) Do() error {
	return r.client.Call("Page.removeScriptToEvaluateOnLoad", r.opts, nil)
}

type SetAutoAttachToCreatedPagesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Controls whether browser will open a new inspector window for connected pages. (experimental)
func (d *Client) SetAutoAttachToCreatedPages() *SetAutoAttachToCreatedPagesRequest {
	return &SetAutoAttachToCreatedPagesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// If true, browser will open a new inspector window for every page created from this one.
func (r *SetAutoAttachToCreatedPagesRequest) AutoAttach(v bool) *SetAutoAttachToCreatedPagesRequest {
	r.opts["autoAttach"] = v
	return r
}

func (r *SetAutoAttachToCreatedPagesRequest) Do() error {
	return r.client.Call("Page.setAutoAttachToCreatedPages", r.opts, nil)
}

type ReloadRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Reloads given page optionally ignoring the cache.
func (d *Client) Reload() *ReloadRequest {
	return &ReloadRequest{opts: make(map[string]interface{}), client: d.Client}
}

// If true, browser cache is ignored (as if the user pressed Shift+refresh). (optional)
func (r *ReloadRequest) IgnoreCache(v bool) *ReloadRequest {
	r.opts["ignoreCache"] = v
	return r
}

// If set, the script will be injected into all frames of the inspected page after reload. (optional)
func (r *ReloadRequest) ScriptToEvaluateOnLoad(v string) *ReloadRequest {
	r.opts["scriptToEvaluateOnLoad"] = v
	return r
}

func (r *ReloadRequest) Do() error {
	return r.client.Call("Page.reload", r.opts, nil)
}

type NavigateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Navigates current page to the given URL.
func (d *Client) Navigate() *NavigateRequest {
	return &NavigateRequest{opts: make(map[string]interface{}), client: d.Client}
}

// URL to navigate the page to.
func (r *NavigateRequest) URL(v string) *NavigateRequest {
	r.opts["url"] = v
	return r
}

// Referrer URL. (optional, experimental)
func (r *NavigateRequest) Referrer(v string) *NavigateRequest {
	r.opts["referrer"] = v
	return r
}

// Intended transition type. (optional, experimental)
func (r *NavigateRequest) TransitionType(v TransitionType) *NavigateRequest {
	r.opts["transitionType"] = v
	return r
}

type NavigateResult struct {
	// Frame id that will be navigated.
	FrameId FrameId `json:"frameId"`
}

func (r *NavigateRequest) Do() (*NavigateResult, error) {
	var result NavigateResult
	err := r.client.Call("Page.navigate", r.opts, &result)
	return &result, err
}

type StopLoadingRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Force the page stop all navigations and pending resource fetches. (experimental)
func (d *Client) StopLoading() *StopLoadingRequest {
	return &StopLoadingRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *StopLoadingRequest) Do() error {
	return r.client.Call("Page.stopLoading", r.opts, nil)
}

type GetNavigationHistoryRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns navigation history for the current page. (experimental)
func (d *Client) GetNavigationHistory() *GetNavigationHistoryRequest {
	return &GetNavigationHistoryRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetNavigationHistoryResult struct {
	// Index of the current navigation history entry.
	CurrentIndex int `json:"currentIndex"`

	// Array of navigation history entries.
	Entries []*NavigationEntry `json:"entries"`
}

func (r *GetNavigationHistoryRequest) Do() (*GetNavigationHistoryResult, error) {
	var result GetNavigationHistoryResult
	err := r.client.Call("Page.getNavigationHistory", r.opts, &result)
	return &result, err
}

type NavigateToHistoryEntryRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Navigates current page to the given history entry. (experimental)
func (d *Client) NavigateToHistoryEntry() *NavigateToHistoryEntryRequest {
	return &NavigateToHistoryEntryRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Unique id of the entry to navigate to.
func (r *NavigateToHistoryEntryRequest) EntryId(v int) *NavigateToHistoryEntryRequest {
	r.opts["entryId"] = v
	return r
}

func (r *NavigateToHistoryEntryRequest) Do() error {
	return r.client.Call("Page.navigateToHistoryEntry", r.opts, nil)
}

type GetCookiesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns all browser cookies. Depending on the backend support, will return detailed cookie information in the <code>cookies</code> field. (experimental)
func (d *Client) GetCookies() *GetCookiesRequest {
	return &GetCookiesRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetCookiesResult struct {
	// Array of cookie objects.
	Cookies []*network.Cookie `json:"cookies"`
}

func (r *GetCookiesRequest) Do() (*GetCookiesResult, error) {
	var result GetCookiesResult
	err := r.client.Call("Page.getCookies", r.opts, &result)
	return &result, err
}

type DeleteCookieRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Deletes browser cookie with given name, domain and path. (experimental)
func (d *Client) DeleteCookie() *DeleteCookieRequest {
	return &DeleteCookieRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Name of the cookie to remove.
func (r *DeleteCookieRequest) CookieName(v string) *DeleteCookieRequest {
	r.opts["cookieName"] = v
	return r
}

// URL to match cooke domain and path.
func (r *DeleteCookieRequest) URL(v string) *DeleteCookieRequest {
	r.opts["url"] = v
	return r
}

func (r *DeleteCookieRequest) Do() error {
	return r.client.Call("Page.deleteCookie", r.opts, nil)
}

type GetResourceTreeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns present frame / resource tree structure. (experimental)
func (d *Client) GetResourceTree() *GetResourceTreeRequest {
	return &GetResourceTreeRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetResourceTreeResult struct {
	// Present frame / resource tree structure.
	FrameTree *FrameResourceTree `json:"frameTree"`
}

func (r *GetResourceTreeRequest) Do() (*GetResourceTreeResult, error) {
	var result GetResourceTreeResult
	err := r.client.Call("Page.getResourceTree", r.opts, &result)
	return &result, err
}

type GetResourceContentRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns content of the given resource. (experimental)
func (d *Client) GetResourceContent() *GetResourceContentRequest {
	return &GetResourceContentRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Frame id to get resource for.
func (r *GetResourceContentRequest) FrameId(v FrameId) *GetResourceContentRequest {
	r.opts["frameId"] = v
	return r
}

// URL of the resource to get content for.
func (r *GetResourceContentRequest) URL(v string) *GetResourceContentRequest {
	r.opts["url"] = v
	return r
}

type GetResourceContentResult struct {
	// Resource content.
	Content string `json:"content"`

	// True, if content was served as base64.
	Base64Encoded bool `json:"base64Encoded"`
}

func (r *GetResourceContentRequest) Do() (*GetResourceContentResult, error) {
	var result GetResourceContentResult
	err := r.client.Call("Page.getResourceContent", r.opts, &result)
	return &result, err
}

type SearchInResourceRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Searches for given string in resource content. (experimental)
func (d *Client) SearchInResource() *SearchInResourceRequest {
	return &SearchInResourceRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Frame id for resource to search in.
func (r *SearchInResourceRequest) FrameId(v FrameId) *SearchInResourceRequest {
	r.opts["frameId"] = v
	return r
}

// URL of the resource to search in.
func (r *SearchInResourceRequest) URL(v string) *SearchInResourceRequest {
	r.opts["url"] = v
	return r
}

// String to search for.
func (r *SearchInResourceRequest) Query(v string) *SearchInResourceRequest {
	r.opts["query"] = v
	return r
}

// If true, search is case sensitive. (optional)
func (r *SearchInResourceRequest) CaseSensitive(v bool) *SearchInResourceRequest {
	r.opts["caseSensitive"] = v
	return r
}

// If true, treats string parameter as regex. (optional)
func (r *SearchInResourceRequest) IsRegex(v bool) *SearchInResourceRequest {
	r.opts["isRegex"] = v
	return r
}

type SearchInResourceResult struct {
	// List of search matches.
	Result []*debugger.SearchMatch `json:"result"`
}

func (r *SearchInResourceRequest) Do() (*SearchInResourceResult, error) {
	var result SearchInResourceResult
	err := r.client.Call("Page.searchInResource", r.opts, &result)
	return &result, err
}

type SetDocumentContentRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets given markup as the document's HTML. (experimental)
func (d *Client) SetDocumentContent() *SetDocumentContentRequest {
	return &SetDocumentContentRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Frame id to set HTML for.
func (r *SetDocumentContentRequest) FrameId(v FrameId) *SetDocumentContentRequest {
	r.opts["frameId"] = v
	return r
}

// HTML content to set.
func (r *SetDocumentContentRequest) Html(v string) *SetDocumentContentRequest {
	r.opts["html"] = v
	return r
}

func (r *SetDocumentContentRequest) Do() error {
	return r.client.Call("Page.setDocumentContent", r.opts, nil)
}

type SetDeviceMetricsOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Overrides the values of device screen dimensions (window.screen.width, window.screen.height, window.innerWidth, window.innerHeight, and "device-width"/"device-height"-related CSS media query results). (experimental)
func (d *Client) SetDeviceMetricsOverride() *SetDeviceMetricsOverrideRequest {
	return &SetDeviceMetricsOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Overriding width value in pixels (minimum 0, maximum 10000000). 0 disables the override.
func (r *SetDeviceMetricsOverrideRequest) Width(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["width"] = v
	return r
}

// Overriding height value in pixels (minimum 0, maximum 10000000). 0 disables the override.
func (r *SetDeviceMetricsOverrideRequest) Height(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["height"] = v
	return r
}

// Overriding device scale factor value. 0 disables the override.
func (r *SetDeviceMetricsOverrideRequest) DeviceScaleFactor(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["deviceScaleFactor"] = v
	return r
}

// Whether to emulate mobile device. This includes viewport meta tag, overlay scrollbars, text autosizing and more.
func (r *SetDeviceMetricsOverrideRequest) Mobile(v bool) *SetDeviceMetricsOverrideRequest {
	r.opts["mobile"] = v
	return r
}

// Whether a view that exceeds the available browser window area should be scaled down to fit.
func (r *SetDeviceMetricsOverrideRequest) FitWindow(v bool) *SetDeviceMetricsOverrideRequest {
	r.opts["fitWindow"] = v
	return r
}

// Scale to apply to resulting view image. Ignored in |fitWindow| mode. (optional)
func (r *SetDeviceMetricsOverrideRequest) Scale(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["scale"] = v
	return r
}

// X offset to shift resulting view image by. Ignored in |fitWindow| mode. (optional)
func (r *SetDeviceMetricsOverrideRequest) OffsetX(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["offsetX"] = v
	return r
}

// Y offset to shift resulting view image by. Ignored in |fitWindow| mode. (optional)
func (r *SetDeviceMetricsOverrideRequest) OffsetY(v float64) *SetDeviceMetricsOverrideRequest {
	r.opts["offsetY"] = v
	return r
}

// Overriding screen width value in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional)
func (r *SetDeviceMetricsOverrideRequest) ScreenWidth(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["screenWidth"] = v
	return r
}

// Overriding screen height value in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional)
func (r *SetDeviceMetricsOverrideRequest) ScreenHeight(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["screenHeight"] = v
	return r
}

// Overriding view X position on screen in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional)
func (r *SetDeviceMetricsOverrideRequest) PositionX(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["positionX"] = v
	return r
}

// Overriding view Y position on screen in pixels (minimum 0, maximum 10000000). Only used for |mobile==true|. (optional)
func (r *SetDeviceMetricsOverrideRequest) PositionY(v int) *SetDeviceMetricsOverrideRequest {
	r.opts["positionY"] = v
	return r
}

// Screen orientation override. (optional)
func (r *SetDeviceMetricsOverrideRequest) ScreenOrientation(v *emulation.ScreenOrientation) *SetDeviceMetricsOverrideRequest {
	r.opts["screenOrientation"] = v
	return r
}

func (r *SetDeviceMetricsOverrideRequest) Do() error {
	return r.client.Call("Page.setDeviceMetricsOverride", r.opts, nil)
}

type ClearDeviceMetricsOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears the overriden device metrics. (experimental)
func (d *Client) ClearDeviceMetricsOverride() *ClearDeviceMetricsOverrideRequest {
	return &ClearDeviceMetricsOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearDeviceMetricsOverrideRequest) Do() error {
	return r.client.Call("Page.clearDeviceMetricsOverride", r.opts, nil)
}

type SetGeolocationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Overrides the Geolocation Position or Error. Omitting any of the parameters emulates position unavailable.
func (d *Client) SetGeolocationOverride() *SetGeolocationOverrideRequest {
	return &SetGeolocationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Mock latitude (optional)
func (r *SetGeolocationOverrideRequest) Latitude(v float64) *SetGeolocationOverrideRequest {
	r.opts["latitude"] = v
	return r
}

// Mock longitude (optional)
func (r *SetGeolocationOverrideRequest) Longitude(v float64) *SetGeolocationOverrideRequest {
	r.opts["longitude"] = v
	return r
}

// Mock accuracy (optional)
func (r *SetGeolocationOverrideRequest) Accuracy(v float64) *SetGeolocationOverrideRequest {
	r.opts["accuracy"] = v
	return r
}

func (r *SetGeolocationOverrideRequest) Do() error {
	return r.client.Call("Page.setGeolocationOverride", r.opts, nil)
}

type ClearGeolocationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears the overriden Geolocation Position and Error.
func (d *Client) ClearGeolocationOverride() *ClearGeolocationOverrideRequest {
	return &ClearGeolocationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearGeolocationOverrideRequest) Do() error {
	return r.client.Call("Page.clearGeolocationOverride", r.opts, nil)
}

type SetDeviceOrientationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Overrides the Device Orientation. (experimental)
func (d *Client) SetDeviceOrientationOverride() *SetDeviceOrientationOverrideRequest {
	return &SetDeviceOrientationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Mock alpha
func (r *SetDeviceOrientationOverrideRequest) Alpha(v float64) *SetDeviceOrientationOverrideRequest {
	r.opts["alpha"] = v
	return r
}

// Mock beta
func (r *SetDeviceOrientationOverrideRequest) Beta(v float64) *SetDeviceOrientationOverrideRequest {
	r.opts["beta"] = v
	return r
}

// Mock gamma
func (r *SetDeviceOrientationOverrideRequest) Gamma(v float64) *SetDeviceOrientationOverrideRequest {
	r.opts["gamma"] = v
	return r
}

func (r *SetDeviceOrientationOverrideRequest) Do() error {
	return r.client.Call("Page.setDeviceOrientationOverride", r.opts, nil)
}

type ClearDeviceOrientationOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears the overridden Device Orientation. (experimental)
func (d *Client) ClearDeviceOrientationOverride() *ClearDeviceOrientationOverrideRequest {
	return &ClearDeviceOrientationOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearDeviceOrientationOverrideRequest) Do() error {
	return r.client.Call("Page.clearDeviceOrientationOverride", r.opts, nil)
}

type SetTouchEmulationEnabledRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Toggles mouse event-based touch event emulation. (experimental)
func (d *Client) SetTouchEmulationEnabled() *SetTouchEmulationEnabledRequest {
	return &SetTouchEmulationEnabledRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether the touch event emulation should be enabled.
func (r *SetTouchEmulationEnabledRequest) Enabled(v bool) *SetTouchEmulationEnabledRequest {
	r.opts["enabled"] = v
	return r
}

// Touch/gesture events configuration. Default: current platform. (optional)
func (r *SetTouchEmulationEnabledRequest) Configuration(v string) *SetTouchEmulationEnabledRequest {
	r.opts["configuration"] = v
	return r
}

func (r *SetTouchEmulationEnabledRequest) Do() error {
	return r.client.Call("Page.setTouchEmulationEnabled", r.opts, nil)
}

type CaptureScreenshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Capture page screenshot. (experimental)
func (d *Client) CaptureScreenshot() *CaptureScreenshotRequest {
	return &CaptureScreenshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Image compression format (defaults to png). (optional)
func (r *CaptureScreenshotRequest) Format(v string) *CaptureScreenshotRequest {
	r.opts["format"] = v
	return r
}

// Compression quality from range [0..100] (jpeg only). (optional)
func (r *CaptureScreenshotRequest) Quality(v int) *CaptureScreenshotRequest {
	r.opts["quality"] = v
	return r
}

// Capture the screenshot from the surface, rather than the view. Defaults to true. (optional, experimental)
func (r *CaptureScreenshotRequest) FromSurface(v bool) *CaptureScreenshotRequest {
	r.opts["fromSurface"] = v
	return r
}

type CaptureScreenshotResult struct {
	// Base64-encoded image data.
	Data string `json:"data"`
}

func (r *CaptureScreenshotRequest) Do() (*CaptureScreenshotResult, error) {
	var result CaptureScreenshotResult
	err := r.client.Call("Page.captureScreenshot", r.opts, &result)
	return &result, err
}

type PrintToPDFRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Print page as PDF. (experimental)
func (d *Client) PrintToPDF() *PrintToPDFRequest {
	return &PrintToPDFRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Paper orientation. Defaults to false. (optional)
func (r *PrintToPDFRequest) Landscape(v bool) *PrintToPDFRequest {
	r.opts["landscape"] = v
	return r
}

// Display header and footer. Defaults to false. (optional)
func (r *PrintToPDFRequest) DisplayHeaderFooter(v bool) *PrintToPDFRequest {
	r.opts["displayHeaderFooter"] = v
	return r
}

// Print background graphics. Defaults to false. (optional)
func (r *PrintToPDFRequest) PrintBackground(v bool) *PrintToPDFRequest {
	r.opts["printBackground"] = v
	return r
}

// Scale of the webpage rendering. Defaults to 1. (optional)
func (r *PrintToPDFRequest) Scale(v float64) *PrintToPDFRequest {
	r.opts["scale"] = v
	return r
}

// Paper width in inches. Defaults to 8.5 inches. (optional)
func (r *PrintToPDFRequest) PaperWidth(v float64) *PrintToPDFRequest {
	r.opts["paperWidth"] = v
	return r
}

// Paper height in inches. Defaults to 11 inches. (optional)
func (r *PrintToPDFRequest) PaperHeight(v float64) *PrintToPDFRequest {
	r.opts["paperHeight"] = v
	return r
}

// Top margin in inches. Defaults to 1cm (~0.4 inches). (optional)
func (r *PrintToPDFRequest) MarginTop(v float64) *PrintToPDFRequest {
	r.opts["marginTop"] = v
	return r
}

// Bottom margin in inches. Defaults to 1cm (~0.4 inches). (optional)
func (r *PrintToPDFRequest) MarginBottom(v float64) *PrintToPDFRequest {
	r.opts["marginBottom"] = v
	return r
}

// Left margin in inches. Defaults to 1cm (~0.4 inches). (optional)
func (r *PrintToPDFRequest) MarginLeft(v float64) *PrintToPDFRequest {
	r.opts["marginLeft"] = v
	return r
}

// Right margin in inches. Defaults to 1cm (~0.4 inches). (optional)
func (r *PrintToPDFRequest) MarginRight(v float64) *PrintToPDFRequest {
	r.opts["marginRight"] = v
	return r
}

// Paper ranges to print, e.g., '1-5, 8, 11-13'. Defaults to the empty string, which means print all pages. (optional)
func (r *PrintToPDFRequest) PageRanges(v string) *PrintToPDFRequest {
	r.opts["pageRanges"] = v
	return r
}

type PrintToPDFResult struct {
	// Base64-encoded pdf data.
	Data string `json:"data"`
}

func (r *PrintToPDFRequest) Do() (*PrintToPDFResult, error) {
	var result PrintToPDFResult
	err := r.client.Call("Page.printToPDF", r.opts, &result)
	return &result, err
}

type StartScreencastRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Starts sending each frame using the <code>screencastFrame</code> event. (experimental)
func (d *Client) StartScreencast() *StartScreencastRequest {
	return &StartScreencastRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Image compression format. (optional)
func (r *StartScreencastRequest) Format(v string) *StartScreencastRequest {
	r.opts["format"] = v
	return r
}

// Compression quality from range [0..100]. (optional)
func (r *StartScreencastRequest) Quality(v int) *StartScreencastRequest {
	r.opts["quality"] = v
	return r
}

// Maximum screenshot width. (optional)
func (r *StartScreencastRequest) MaxWidth(v int) *StartScreencastRequest {
	r.opts["maxWidth"] = v
	return r
}

// Maximum screenshot height. (optional)
func (r *StartScreencastRequest) MaxHeight(v int) *StartScreencastRequest {
	r.opts["maxHeight"] = v
	return r
}

// Send every n-th frame. (optional)
func (r *StartScreencastRequest) EveryNthFrame(v int) *StartScreencastRequest {
	r.opts["everyNthFrame"] = v
	return r
}

func (r *StartScreencastRequest) Do() error {
	return r.client.Call("Page.startScreencast", r.opts, nil)
}

type StopScreencastRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Stops sending each frame in the <code>screencastFrame</code>. (experimental)
func (d *Client) StopScreencast() *StopScreencastRequest {
	return &StopScreencastRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *StopScreencastRequest) Do() error {
	return r.client.Call("Page.stopScreencast", r.opts, nil)
}

type ScreencastFrameAckRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Acknowledges that a screencast frame has been received by the frontend. (experimental)
func (d *Client) ScreencastFrameAck() *ScreencastFrameAckRequest {
	return &ScreencastFrameAckRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Frame number.
func (r *ScreencastFrameAckRequest) SessionId(v int) *ScreencastFrameAckRequest {
	r.opts["sessionId"] = v
	return r
}

func (r *ScreencastFrameAckRequest) Do() error {
	return r.client.Call("Page.screencastFrameAck", r.opts, nil)
}

type HandleJavaScriptDialogRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Accepts or dismisses a JavaScript initiated dialog (alert, confirm, prompt, or onbeforeunload).
func (d *Client) HandleJavaScriptDialog() *HandleJavaScriptDialogRequest {
	return &HandleJavaScriptDialogRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether to accept or dismiss the dialog.
func (r *HandleJavaScriptDialogRequest) Accept(v bool) *HandleJavaScriptDialogRequest {
	r.opts["accept"] = v
	return r
}

// The text to enter into the dialog prompt before accepting. Used only if this is a prompt dialog. (optional)
func (r *HandleJavaScriptDialogRequest) PromptText(v string) *HandleJavaScriptDialogRequest {
	r.opts["promptText"] = v
	return r
}

func (r *HandleJavaScriptDialogRequest) Do() error {
	return r.client.Call("Page.handleJavaScriptDialog", r.opts, nil)
}

type GetAppManifestRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// (experimental)
func (d *Client) GetAppManifest() *GetAppManifestRequest {
	return &GetAppManifestRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetAppManifestResult struct {
	// Manifest location.
	URL string `json:"url"`

	Errors []*AppManifestError `json:"errors"`

	// Manifest content. (optional)
	Data string `json:"data"`
}

func (r *GetAppManifestRequest) Do() (*GetAppManifestResult, error) {
	var result GetAppManifestResult
	err := r.client.Call("Page.getAppManifest", r.opts, &result)
	return &result, err
}

type RequestAppBannerRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// (experimental)
func (d *Client) RequestAppBanner() *RequestAppBannerRequest {
	return &RequestAppBannerRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *RequestAppBannerRequest) Do() error {
	return r.client.Call("Page.requestAppBanner", r.opts, nil)
}

type SetControlNavigationsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Toggles navigation throttling which allows programatic control over navigation and redirect response. (experimental)
func (d *Client) SetControlNavigations() *SetControlNavigationsRequest {
	return &SetControlNavigationsRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetControlNavigationsRequest) Enabled(v bool) *SetControlNavigationsRequest {
	r.opts["enabled"] = v
	return r
}

func (r *SetControlNavigationsRequest) Do() error {
	return r.client.Call("Page.setControlNavigations", r.opts, nil)
}

type ProcessNavigationRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Should be sent in response to a navigationRequested or a redirectRequested event, telling the browser how to handle the navigation. (experimental)
func (d *Client) ProcessNavigation() *ProcessNavigationRequest {
	return &ProcessNavigationRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ProcessNavigationRequest) Response(v NavigationResponse) *ProcessNavigationRequest {
	r.opts["response"] = v
	return r
}

func (r *ProcessNavigationRequest) NavigationId(v int) *ProcessNavigationRequest {
	r.opts["navigationId"] = v
	return r
}

func (r *ProcessNavigationRequest) Do() error {
	return r.client.Call("Page.processNavigation", r.opts, nil)
}

type GetLayoutMetricsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns metrics relating to the layouting of the page, such as viewport bounds/scale. (experimental)
func (d *Client) GetLayoutMetrics() *GetLayoutMetricsRequest {
	return &GetLayoutMetricsRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetLayoutMetricsResult struct {
	// Metrics relating to the layout viewport.
	LayoutViewport *LayoutViewport `json:"layoutViewport"`

	// Metrics relating to the visual viewport.
	VisualViewport *VisualViewport `json:"visualViewport"`

	// Size of scrollable area.
	ContentSize *dom.Rect `json:"contentSize"`
}

func (r *GetLayoutMetricsRequest) Do() (*GetLayoutMetricsResult, error) {
	var result GetLayoutMetricsResult
	err := r.client.Call("Page.getLayoutMetrics", r.opts, &result)
	return &result, err
}

type CreateIsolatedWorldRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Creates an isolated world for the given frame. (experimental)
func (d *Client) CreateIsolatedWorld() *CreateIsolatedWorldRequest {
	return &CreateIsolatedWorldRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Id of the frame in which the isolated world should be created.
func (r *CreateIsolatedWorldRequest) FrameId(v FrameId) *CreateIsolatedWorldRequest {
	r.opts["frameId"] = v
	return r
}

// An optional name which is reported in the Execution Context. (optional)
func (r *CreateIsolatedWorldRequest) WorldName(v string) *CreateIsolatedWorldRequest {
	r.opts["worldName"] = v
	return r
}

// Whether or not universal access should be granted to the isolated world. This is a powerful option, use with caution. (optional)
func (r *CreateIsolatedWorldRequest) GrantUniveralAccess(v bool) *CreateIsolatedWorldRequest {
	r.opts["grantUniveralAccess"] = v
	return r
}

func (r *CreateIsolatedWorldRequest) Do() error {
	return r.client.Call("Page.createIsolatedWorld", r.opts, nil)
}

func init() {
	rpc.EventTypes["Page.domContentEventFired"] = func() interface{} { return new(DomContentEventFiredEvent) }
	rpc.EventTypes["Page.loadEventFired"] = func() interface{} { return new(LoadEventFiredEvent) }
	rpc.EventTypes["Page.frameAttached"] = func() interface{} { return new(FrameAttachedEvent) }
	rpc.EventTypes["Page.frameNavigated"] = func() interface{} { return new(FrameNavigatedEvent) }
	rpc.EventTypes["Page.frameDetached"] = func() interface{} { return new(FrameDetachedEvent) }
	rpc.EventTypes["Page.frameStartedLoading"] = func() interface{} { return new(FrameStartedLoadingEvent) }
	rpc.EventTypes["Page.frameStoppedLoading"] = func() interface{} { return new(FrameStoppedLoadingEvent) }
	rpc.EventTypes["Page.frameScheduledNavigation"] = func() interface{} { return new(FrameScheduledNavigationEvent) }
	rpc.EventTypes["Page.frameClearedScheduledNavigation"] = func() interface{} { return new(FrameClearedScheduledNavigationEvent) }
	rpc.EventTypes["Page.frameResized"] = func() interface{} { return new(FrameResizedEvent) }
	rpc.EventTypes["Page.javascriptDialogOpening"] = func() interface{} { return new(JavascriptDialogOpeningEvent) }
	rpc.EventTypes["Page.javascriptDialogClosed"] = func() interface{} { return new(JavascriptDialogClosedEvent) }
	rpc.EventTypes["Page.screencastFrame"] = func() interface{} { return new(ScreencastFrameEvent) }
	rpc.EventTypes["Page.screencastVisibilityChanged"] = func() interface{} { return new(ScreencastVisibilityChangedEvent) }
	rpc.EventTypes["Page.interstitialShown"] = func() interface{} { return new(InterstitialShownEvent) }
	rpc.EventTypes["Page.interstitialHidden"] = func() interface{} { return new(InterstitialHiddenEvent) }
	rpc.EventTypes["Page.navigationRequested"] = func() interface{} { return new(NavigationRequestedEvent) }
}

type DomContentEventFiredEvent struct {
	Timestamp float64 `json:"timestamp"`
}

type LoadEventFiredEvent struct {
	Timestamp float64 `json:"timestamp"`
}

// Fired when frame has been attached to its parent.
type FrameAttachedEvent struct {
	// Id of the frame that has been attached.
	FrameId FrameId `json:"frameId"`

	// Parent frame identifier.
	ParentFrameId FrameId `json:"parentFrameId"`

	// JavaScript stack trace of when frame was attached, only set if frame initiated from script. (optional, experimental)
	Stack *runtime.StackTrace `json:"stack"`
}

// Fired once navigation of the frame has completed. Frame is now associated with the new loader.
type FrameNavigatedEvent struct {
	// Frame object.
	Frame *Frame `json:"frame"`
}

// Fired when frame has been detached from its parent.
type FrameDetachedEvent struct {
	// Id of the frame that has been detached.
	FrameId FrameId `json:"frameId"`
}

// Fired when frame has started loading. (experimental)
type FrameStartedLoadingEvent struct {
	// Id of the frame that has started loading.
	FrameId FrameId `json:"frameId"`
}

// Fired when frame has stopped loading. (experimental)
type FrameStoppedLoadingEvent struct {
	// Id of the frame that has stopped loading.
	FrameId FrameId `json:"frameId"`
}

// Fired when frame schedules a potential navigation. (experimental)
type FrameScheduledNavigationEvent struct {
	// Id of the frame that has scheduled a navigation.
	FrameId FrameId `json:"frameId"`

	// Delay (in seconds) until the navigation is scheduled to begin. The navigation is not guaranteed to start.
	Delay float64 `json:"delay"`
}

// Fired when frame no longer has a scheduled navigation. (experimental)
type FrameClearedScheduledNavigationEvent struct {
	// Id of the frame that has cleared its scheduled navigation.
	FrameId FrameId `json:"frameId"`
}

// (experimental)
type FrameResizedEvent struct {
}

// Fired when a JavaScript initiated dialog (alert, confirm, prompt, or onbeforeunload) is about to open.
type JavascriptDialogOpeningEvent struct {
	// Message that will be displayed by the dialog.
	Message string `json:"message"`

	// Dialog type.
	Type DialogType `json:"type"`
}

// Fired when a JavaScript initiated dialog (alert, confirm, prompt, or onbeforeunload) has been closed.
type JavascriptDialogClosedEvent struct {
	// Whether dialog was confirmed.
	Result bool `json:"result"`
}

// Compressed image data requested by the <code>startScreencast</code>. (experimental)
type ScreencastFrameEvent struct {
	// Base64-encoded compressed image.
	Data string `json:"data"`

	// Screencast frame metadata.
	Metadata *ScreencastFrameMetadata `json:"metadata"`

	// Frame number.
	SessionId int `json:"sessionId"`
}

// Fired when the page with currently enabled screencast was shown or hidden </code>. (experimental)
type ScreencastVisibilityChangedEvent struct {
	// True if the page is visible.
	Visible bool `json:"visible"`
}

// Fired when interstitial page was shown
type InterstitialShownEvent struct {
}

// Fired when interstitial page was hidden
type InterstitialHiddenEvent struct {
}

// Fired when a navigation is started if navigation throttles are enabled.  The navigation will be deferred until processNavigation is called.
type NavigationRequestedEvent struct {
	// Whether the navigation is taking place in the main frame or in a subframe.
	IsInMainFrame bool `json:"isInMainFrame"`

	// Whether the navigation has encountered a server redirect or not.
	IsRedirect bool `json:"isRedirect"`

	NavigationId int `json:"navigationId"`

	// URL of requested navigation.
	URL string `json:"url"`
}

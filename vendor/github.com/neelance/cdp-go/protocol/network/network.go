// Network domain allows tracking network activities of the page. It exposes information about http, file, data and other requests and responses, their headers, bodies, timing, etc.
package network

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/runtime"
	"github.com/neelance/cdp-go/protocol/security"
)

// Network domain allows tracking network activities of the page. It exposes information about http, file, data and other requests and responses, their headers, bodies, timing, etc.
type Client struct {
	*rpc.Client
}

// Unique loader identifier.

type LoaderId string

// Unique request identifier.

type RequestId string

// Unique intercepted request identifier.

type InterceptionId string

// Network level fetch failure reason.

type ErrorReason string

// Number of seconds since epoch.

type Timestamp float64

// Request / response headers as keys / values of JSON object.

type Headers struct {
}

// Loading priority of a resource request.

type ConnectionType string

// Represents the cookie's 'SameSite' status: https://tools.ietf.org/html/draft-west-first-party-cookies

type CookieSameSite string

// Timing information for the request.

type ResourceTiming struct {
	// Timing's requestTime is a baseline in seconds, while the other numbers are ticks in milliseconds relatively to this requestTime.
	RequestTime float64 `json:"requestTime"`

	// Started resolving proxy.
	ProxyStart float64 `json:"proxyStart"`

	// Finished resolving proxy.
	ProxyEnd float64 `json:"proxyEnd"`

	// Started DNS address resolve.
	DnsStart float64 `json:"dnsStart"`

	// Finished DNS address resolve.
	DnsEnd float64 `json:"dnsEnd"`

	// Started connecting to the remote host.
	ConnectStart float64 `json:"connectStart"`

	// Connected to the remote host.
	ConnectEnd float64 `json:"connectEnd"`

	// Started SSL handshake.
	SslStart float64 `json:"sslStart"`

	// Finished SSL handshake.
	SslEnd float64 `json:"sslEnd"`

	// Started running ServiceWorker.
	WorkerStart float64 `json:"workerStart"`

	// Finished Starting ServiceWorker.
	WorkerReady float64 `json:"workerReady"`

	// Started sending request.
	SendStart float64 `json:"sendStart"`

	// Finished sending request.
	SendEnd float64 `json:"sendEnd"`

	// Time the server started pushing request.
	PushStart float64 `json:"pushStart"`

	// Time the server finished pushing request.
	PushEnd float64 `json:"pushEnd"`

	// Finished receiving response headers.
	ReceiveHeadersEnd float64 `json:"receiveHeadersEnd"`
}

// Loading priority of a resource request.

type ResourcePriority string

// HTTP request data.

type Request struct {
	// Request URL.
	URL string `json:"url"`

	// HTTP request method.
	Method string `json:"method"`

	// HTTP request headers.
	Headers *Headers `json:"headers"`

	// HTTP POST request data. (optional)
	PostData string `json:"postData,omitempty"`

	// The mixed content status of the request, as defined in http://www.w3.org/TR/mixed-content/ (optional)
	MixedContentType string `json:"mixedContentType,omitempty"`

	// Priority of the resource request at the time request is sent.
	InitialPriority ResourcePriority `json:"initialPriority"`

	// The referrer policy of the request, as defined in https://www.w3.org/TR/referrer-policy/
	ReferrerPolicy string `json:"referrerPolicy"`

	// Whether is loaded via link preload. (optional)
	IsLinkPreload bool `json:"isLinkPreload,omitempty"`
}

// Details of a signed certificate timestamp (SCT).

type SignedCertificateTimestamp struct {
	// Validation status.
	Status string `json:"status"`

	// Origin.
	Origin string `json:"origin"`

	// Log name / description.
	LogDescription string `json:"logDescription"`

	// Log ID.
	LogId string `json:"logId"`

	// Issuance date.
	Timestamp Timestamp `json:"timestamp"`

	// Hash algorithm.
	HashAlgorithm string `json:"hashAlgorithm"`

	// Signature algorithm.
	SignatureAlgorithm string `json:"signatureAlgorithm"`

	// Signature data.
	SignatureData string `json:"signatureData"`
}

// Security details about a request.

type SecurityDetails struct {
	// Protocol name (e.g. "TLS 1.2" or "QUIC").
	Protocol string `json:"protocol"`

	// Key Exchange used by the connection, or the empty string if not applicable.
	KeyExchange string `json:"keyExchange"`

	// (EC)DH group used by the connection, if applicable. (optional)
	KeyExchangeGroup string `json:"keyExchangeGroup,omitempty"`

	// Cipher name.
	Cipher string `json:"cipher"`

	// TLS MAC. Note that AEAD ciphers do not have separate MACs. (optional)
	Mac string `json:"mac,omitempty"`

	// Certificate ID value.
	CertificateId security.CertificateId `json:"certificateId"`

	// Certificate subject name.
	SubjectName string `json:"subjectName"`

	// Subject Alternative Name (SAN) DNS names and IP addresses.
	SanList []string `json:"sanList"`

	// Name of the issuing CA.
	Issuer string `json:"issuer"`

	// Certificate valid from date.
	ValidFrom Timestamp `json:"validFrom"`

	// Certificate valid to (expiration) date
	ValidTo Timestamp `json:"validTo"`

	// List of signed certificate timestamps (SCTs).
	SignedCertificateTimestampList []*SignedCertificateTimestamp `json:"signedCertificateTimestampList"`
}

// The reason why request was blocked. (experimental)

type BlockedReason string

// HTTP response data.

type Response struct {
	// Response URL. This URL can be different from CachedResource.url in case of redirect.
	URL string `json:"url"`

	// HTTP response status code.
	Status float64 `json:"status"`

	// HTTP response status text.
	StatusText string `json:"statusText"`

	// HTTP response headers.
	Headers *Headers `json:"headers"`

	// HTTP response headers text. (optional)
	HeadersText string `json:"headersText,omitempty"`

	// Resource mimeType as determined by the browser.
	MimeType string `json:"mimeType"`

	// Refined HTTP request headers that were actually transmitted over the network. (optional)
	RequestHeaders *Headers `json:"requestHeaders,omitempty"`

	// HTTP request headers text. (optional)
	RequestHeadersText string `json:"requestHeadersText,omitempty"`

	// Specifies whether physical connection was actually reused for this request.
	ConnectionReused bool `json:"connectionReused"`

	// Physical connection id that was actually used for this request.
	ConnectionId float64 `json:"connectionId"`

	// Remote IP address. (optional, experimental)
	RemoteIPAddress string `json:"remoteIPAddress,omitempty"`

	// Remote port. (optional, experimental)
	RemotePort int `json:"remotePort,omitempty"`

	// Specifies that the request was served from the disk cache. (optional)
	FromDiskCache bool `json:"fromDiskCache,omitempty"`

	// Specifies that the request was served from the ServiceWorker. (optional)
	FromServiceWorker bool `json:"fromServiceWorker,omitempty"`

	// Total number of bytes received for this request so far.
	EncodedDataLength float64 `json:"encodedDataLength"`

	// Timing information for the given request. (optional)
	Timing *ResourceTiming `json:"timing,omitempty"`

	// Protocol used to fetch this request. (optional)
	Protocol string `json:"protocol,omitempty"`

	// Security state of the request resource.
	SecurityState security.SecurityState `json:"securityState"`

	// Security details for the request. (optional)
	SecurityDetails *SecurityDetails `json:"securityDetails,omitempty"`
}

// WebSocket request data. (experimental)

type WebSocketRequest struct {
	// HTTP request headers.
	Headers *Headers `json:"headers"`
}

// WebSocket response data. (experimental)

type WebSocketResponse struct {
	// HTTP response status code.
	Status float64 `json:"status"`

	// HTTP response status text.
	StatusText string `json:"statusText"`

	// HTTP response headers.
	Headers *Headers `json:"headers"`

	// HTTP response headers text. (optional)
	HeadersText string `json:"headersText,omitempty"`

	// HTTP request headers. (optional)
	RequestHeaders *Headers `json:"requestHeaders,omitempty"`

	// HTTP request headers text. (optional)
	RequestHeadersText string `json:"requestHeadersText,omitempty"`
}

// WebSocket frame data. (experimental)

type WebSocketFrame struct {
	// WebSocket frame opcode.
	Opcode float64 `json:"opcode"`

	// WebSocke frame mask.
	Mask bool `json:"mask"`

	// WebSocke frame payload data.
	PayloadData string `json:"payloadData"`
}

// Information about the cached resource.

type CachedResource struct {
	// Resource URL. This is the url of the original network request.
	URL string `json:"url"`

	// Type of this resource.
	Type string `json:"type"`

	// Cached response data. (optional)
	Response *Response `json:"response,omitempty"`

	// Cached response body size.
	BodySize float64 `json:"bodySize"`
}

// Information about the request initiator.

type Initiator struct {
	// Type of this initiator.
	Type string `json:"type"`

	// Initiator JavaScript stack trace, set for Script only. (optional)
	Stack *runtime.StackTrace `json:"stack,omitempty"`

	// Initiator URL, set for Parser type only. (optional)
	URL string `json:"url,omitempty"`

	// Initiator line number, set for Parser type only (0-based). (optional)
	LineNumber float64 `json:"lineNumber,omitempty"`
}

// Cookie object (experimental)

type Cookie struct {
	// Cookie name.
	Name string `json:"name"`

	// Cookie value.
	Value string `json:"value"`

	// Cookie domain.
	Domain string `json:"domain"`

	// Cookie path.
	Path string `json:"path"`

	// Cookie expiration date as the number of seconds since the UNIX epoch.
	Expires float64 `json:"expires"`

	// Cookie size.
	Size int `json:"size"`

	// True if cookie is http-only.
	HttpOnly bool `json:"httpOnly"`

	// True if cookie is secure.
	Secure bool `json:"secure"`

	// True in case of session cookie.
	Session bool `json:"session"`

	// Cookie SameSite type. (optional)
	SameSite CookieSameSite `json:"sameSite,omitempty"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables network tracking, network events will now be delivered to the client.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Buffer size in bytes to use when preserving network payloads (XHRs, etc). (optional, experimental)
func (r *EnableRequest) MaxTotalBufferSize(v int) *EnableRequest {
	r.opts["maxTotalBufferSize"] = v
	return r
}

// Per-resource buffer size in bytes to use when preserving network payloads (XHRs, etc). (optional, experimental)
func (r *EnableRequest) MaxResourceBufferSize(v int) *EnableRequest {
	r.opts["maxResourceBufferSize"] = v
	return r
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Network.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables network tracking, prevents network events from being sent to the client.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Network.disable", r.opts, nil)
}

type SetUserAgentOverrideRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Allows overriding user agent with the given string.
func (d *Client) SetUserAgentOverride() *SetUserAgentOverrideRequest {
	return &SetUserAgentOverrideRequest{opts: make(map[string]interface{}), client: d.Client}
}

// User agent to use.
func (r *SetUserAgentOverrideRequest) UserAgent(v string) *SetUserAgentOverrideRequest {
	r.opts["userAgent"] = v
	return r
}

func (r *SetUserAgentOverrideRequest) Do() error {
	return r.client.Call("Network.setUserAgentOverride", r.opts, nil)
}

type SetExtraHTTPHeadersRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Specifies whether to always send extra HTTP headers with the requests from this page.
func (d *Client) SetExtraHTTPHeaders() *SetExtraHTTPHeadersRequest {
	return &SetExtraHTTPHeadersRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Map with extra HTTP headers.
func (r *SetExtraHTTPHeadersRequest) Headers(v *Headers) *SetExtraHTTPHeadersRequest {
	r.opts["headers"] = v
	return r
}

func (r *SetExtraHTTPHeadersRequest) Do() error {
	return r.client.Call("Network.setExtraHTTPHeaders", r.opts, nil)
}

type GetResponseBodyRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns content served for the given request.
func (d *Client) GetResponseBody() *GetResponseBodyRequest {
	return &GetResponseBodyRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the network request to get content for.
func (r *GetResponseBodyRequest) RequestId(v RequestId) *GetResponseBodyRequest {
	r.opts["requestId"] = v
	return r
}

type GetResponseBodyResult struct {
	// Response body.
	Body string `json:"body"`

	// True, if content was sent as base64.
	Base64Encoded bool `json:"base64Encoded"`
}

func (r *GetResponseBodyRequest) Do() (*GetResponseBodyResult, error) {
	var result GetResponseBodyResult
	err := r.client.Call("Network.getResponseBody", r.opts, &result)
	return &result, err
}

type SetBlockedURLsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Blocks URLs from loading. (experimental)
func (d *Client) SetBlockedURLs() *SetBlockedURLsRequest {
	return &SetBlockedURLsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// URL patterns to block. Wildcards ('*') are allowed.
func (r *SetBlockedURLsRequest) Urls(v []string) *SetBlockedURLsRequest {
	r.opts["urls"] = v
	return r
}

func (r *SetBlockedURLsRequest) Do() error {
	return r.client.Call("Network.setBlockedURLs", r.opts, nil)
}

type ReplayXHRRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// This method sends a new XMLHttpRequest which is identical to the original one. The following parameters should be identical: method, url, async, request body, extra headers, withCredentials attribute, user, password. (experimental)
func (d *Client) ReplayXHR() *ReplayXHRRequest {
	return &ReplayXHRRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of XHR to replay.
func (r *ReplayXHRRequest) RequestId(v RequestId) *ReplayXHRRequest {
	r.opts["requestId"] = v
	return r
}

func (r *ReplayXHRRequest) Do() error {
	return r.client.Call("Network.replayXHR", r.opts, nil)
}

type CanClearBrowserCacheRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Tells whether clearing browser cache is supported.
func (d *Client) CanClearBrowserCache() *CanClearBrowserCacheRequest {
	return &CanClearBrowserCacheRequest{opts: make(map[string]interface{}), client: d.Client}
}

type CanClearBrowserCacheResult struct {
	// True if browser cache can be cleared.
	Result bool `json:"result"`
}

func (r *CanClearBrowserCacheRequest) Do() (*CanClearBrowserCacheResult, error) {
	var result CanClearBrowserCacheResult
	err := r.client.Call("Network.canClearBrowserCache", r.opts, &result)
	return &result, err
}

type ClearBrowserCacheRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears browser cache.
func (d *Client) ClearBrowserCache() *ClearBrowserCacheRequest {
	return &ClearBrowserCacheRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearBrowserCacheRequest) Do() error {
	return r.client.Call("Network.clearBrowserCache", r.opts, nil)
}

type CanClearBrowserCookiesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Tells whether clearing browser cookies is supported.
func (d *Client) CanClearBrowserCookies() *CanClearBrowserCookiesRequest {
	return &CanClearBrowserCookiesRequest{opts: make(map[string]interface{}), client: d.Client}
}

type CanClearBrowserCookiesResult struct {
	// True if browser cookies can be cleared.
	Result bool `json:"result"`
}

func (r *CanClearBrowserCookiesRequest) Do() (*CanClearBrowserCookiesResult, error) {
	var result CanClearBrowserCookiesResult
	err := r.client.Call("Network.canClearBrowserCookies", r.opts, &result)
	return &result, err
}

type ClearBrowserCookiesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears browser cookies.
func (d *Client) ClearBrowserCookies() *ClearBrowserCookiesRequest {
	return &ClearBrowserCookiesRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearBrowserCookiesRequest) Do() error {
	return r.client.Call("Network.clearBrowserCookies", r.opts, nil)
}

type GetCookiesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns all browser cookies for the current URL. Depending on the backend support, will return detailed cookie information in the <code>cookies</code> field. (experimental)
func (d *Client) GetCookies() *GetCookiesRequest {
	return &GetCookiesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The list of URLs for which applicable cookies will be fetched (optional)
func (r *GetCookiesRequest) Urls(v []string) *GetCookiesRequest {
	r.opts["urls"] = v
	return r
}

type GetCookiesResult struct {
	// Array of cookie objects.
	Cookies []*Cookie `json:"cookies"`
}

func (r *GetCookiesRequest) Do() (*GetCookiesResult, error) {
	var result GetCookiesResult
	err := r.client.Call("Network.getCookies", r.opts, &result)
	return &result, err
}

type GetAllCookiesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns all browser cookies. Depending on the backend support, will return detailed cookie information in the <code>cookies</code> field. (experimental)
func (d *Client) GetAllCookies() *GetAllCookiesRequest {
	return &GetAllCookiesRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetAllCookiesResult struct {
	// Array of cookie objects.
	Cookies []*Cookie `json:"cookies"`
}

func (r *GetAllCookiesRequest) Do() (*GetAllCookiesResult, error) {
	var result GetAllCookiesResult
	err := r.client.Call("Network.getAllCookies", r.opts, &result)
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
	return r.client.Call("Network.deleteCookie", r.opts, nil)
}

type SetCookieRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets a cookie with the given cookie data; may overwrite equivalent cookies if they exist. (experimental)
func (d *Client) SetCookie() *SetCookieRequest {
	return &SetCookieRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The request-URI to associate with the setting of the cookie. This value can affect the default domain and path values of the created cookie.
func (r *SetCookieRequest) URL(v string) *SetCookieRequest {
	r.opts["url"] = v
	return r
}

// The name of the cookie.
func (r *SetCookieRequest) Name(v string) *SetCookieRequest {
	r.opts["name"] = v
	return r
}

// The value of the cookie.
func (r *SetCookieRequest) Value(v string) *SetCookieRequest {
	r.opts["value"] = v
	return r
}

// If omitted, the cookie becomes a host-only cookie. (optional)
func (r *SetCookieRequest) Domain(v string) *SetCookieRequest {
	r.opts["domain"] = v
	return r
}

// Defaults to the path portion of the url parameter. (optional)
func (r *SetCookieRequest) Path(v string) *SetCookieRequest {
	r.opts["path"] = v
	return r
}

// Defaults ot false. (optional)
func (r *SetCookieRequest) Secure(v bool) *SetCookieRequest {
	r.opts["secure"] = v
	return r
}

// Defaults to false. (optional)
func (r *SetCookieRequest) HttpOnly(v bool) *SetCookieRequest {
	r.opts["httpOnly"] = v
	return r
}

// Defaults to browser default behavior. (optional)
func (r *SetCookieRequest) SameSite(v CookieSameSite) *SetCookieRequest {
	r.opts["sameSite"] = v
	return r
}

// If omitted, the cookie becomes a session cookie. (optional)
func (r *SetCookieRequest) ExpirationDate(v Timestamp) *SetCookieRequest {
	r.opts["expirationDate"] = v
	return r
}

type SetCookieResult struct {
	// True if successfully set cookie.
	Success bool `json:"success"`
}

func (r *SetCookieRequest) Do() (*SetCookieResult, error) {
	var result SetCookieResult
	err := r.client.Call("Network.setCookie", r.opts, &result)
	return &result, err
}

type CanEmulateNetworkConditionsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Tells whether emulation of network conditions is supported. (experimental)
func (d *Client) CanEmulateNetworkConditions() *CanEmulateNetworkConditionsRequest {
	return &CanEmulateNetworkConditionsRequest{opts: make(map[string]interface{}), client: d.Client}
}

type CanEmulateNetworkConditionsResult struct {
	// True if emulation of network conditions is supported.
	Result bool `json:"result"`
}

func (r *CanEmulateNetworkConditionsRequest) Do() (*CanEmulateNetworkConditionsResult, error) {
	var result CanEmulateNetworkConditionsResult
	err := r.client.Call("Network.canEmulateNetworkConditions", r.opts, &result)
	return &result, err
}

type EmulateNetworkConditionsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Activates emulation of network conditions.
func (d *Client) EmulateNetworkConditions() *EmulateNetworkConditionsRequest {
	return &EmulateNetworkConditionsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// True to emulate internet disconnection.
func (r *EmulateNetworkConditionsRequest) Offline(v bool) *EmulateNetworkConditionsRequest {
	r.opts["offline"] = v
	return r
}

// Additional latency (ms).
func (r *EmulateNetworkConditionsRequest) Latency(v float64) *EmulateNetworkConditionsRequest {
	r.opts["latency"] = v
	return r
}

// Maximal aggregated download throughput.
func (r *EmulateNetworkConditionsRequest) DownloadThroughput(v float64) *EmulateNetworkConditionsRequest {
	r.opts["downloadThroughput"] = v
	return r
}

// Maximal aggregated upload throughput.
func (r *EmulateNetworkConditionsRequest) UploadThroughput(v float64) *EmulateNetworkConditionsRequest {
	r.opts["uploadThroughput"] = v
	return r
}

// Connection type if known. (optional)
func (r *EmulateNetworkConditionsRequest) ConnectionType(v ConnectionType) *EmulateNetworkConditionsRequest {
	r.opts["connectionType"] = v
	return r
}

func (r *EmulateNetworkConditionsRequest) Do() error {
	return r.client.Call("Network.emulateNetworkConditions", r.opts, nil)
}

type SetCacheDisabledRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Toggles ignoring cache for each request. If <code>true</code>, cache will not be used.
func (d *Client) SetCacheDisabled() *SetCacheDisabledRequest {
	return &SetCacheDisabledRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Cache disabled state.
func (r *SetCacheDisabledRequest) CacheDisabled(v bool) *SetCacheDisabledRequest {
	r.opts["cacheDisabled"] = v
	return r
}

func (r *SetCacheDisabledRequest) Do() error {
	return r.client.Call("Network.setCacheDisabled", r.opts, nil)
}

type SetBypassServiceWorkerRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Toggles ignoring of service worker for each request. (experimental)
func (d *Client) SetBypassServiceWorker() *SetBypassServiceWorkerRequest {
	return &SetBypassServiceWorkerRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Bypass service worker and load from network.
func (r *SetBypassServiceWorkerRequest) Bypass(v bool) *SetBypassServiceWorkerRequest {
	r.opts["bypass"] = v
	return r
}

func (r *SetBypassServiceWorkerRequest) Do() error {
	return r.client.Call("Network.setBypassServiceWorker", r.opts, nil)
}

type SetDataSizeLimitsForTestRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// For testing. (experimental)
func (d *Client) SetDataSizeLimitsForTest() *SetDataSizeLimitsForTestRequest {
	return &SetDataSizeLimitsForTestRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Maximum total buffer size.
func (r *SetDataSizeLimitsForTestRequest) MaxTotalSize(v int) *SetDataSizeLimitsForTestRequest {
	r.opts["maxTotalSize"] = v
	return r
}

// Maximum per-resource size.
func (r *SetDataSizeLimitsForTestRequest) MaxResourceSize(v int) *SetDataSizeLimitsForTestRequest {
	r.opts["maxResourceSize"] = v
	return r
}

func (r *SetDataSizeLimitsForTestRequest) Do() error {
	return r.client.Call("Network.setDataSizeLimitsForTest", r.opts, nil)
}

type GetCertificateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns the DER-encoded certificate. (experimental)
func (d *Client) GetCertificate() *GetCertificateRequest {
	return &GetCertificateRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Origin to get certificate for.
func (r *GetCertificateRequest) Origin(v string) *GetCertificateRequest {
	r.opts["origin"] = v
	return r
}

type GetCertificateResult struct {
	TableNames []string `json:"tableNames"`
}

func (r *GetCertificateRequest) Do() (*GetCertificateResult, error) {
	var result GetCertificateResult
	err := r.client.Call("Network.getCertificate", r.opts, &result)
	return &result, err
}

type EnableRequestInterceptionRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// (experimental)
func (d *Client) EnableRequestInterception() *EnableRequestInterceptionRequest {
	return &EnableRequestInterceptionRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whether or not HTTP requests should be intercepted and Network.requestIntercepted events sent.
func (r *EnableRequestInterceptionRequest) Enabled(v bool) *EnableRequestInterceptionRequest {
	r.opts["enabled"] = v
	return r
}

func (r *EnableRequestInterceptionRequest) Do() error {
	return r.client.Call("Network.enableRequestInterception", r.opts, nil)
}

type ContinueInterceptedRequestRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Response to Network.requestIntercepted which either modifies the request to continue with any modifications, or blocks it, or completes it with the provided response bytes. If a network fetch occurs as a result which encounters a redirect an additional Network.requestIntercepted event will be sent with the same InterceptionId. (experimental)
func (d *Client) ContinueInterceptedRequest() *ContinueInterceptedRequestRequest {
	return &ContinueInterceptedRequestRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ContinueInterceptedRequestRequest) InterceptionId(v InterceptionId) *ContinueInterceptedRequestRequest {
	r.opts["interceptionId"] = v
	return r
}

// If set this causes the request to fail with the given reason. (optional)
func (r *ContinueInterceptedRequestRequest) ErrorReason(v ErrorReason) *ContinueInterceptedRequestRequest {
	r.opts["errorReason"] = v
	return r
}

// If set the requests completes using with the provided base64 encoded raw response, including HTTP status line and headers etc... (optional)
func (r *ContinueInterceptedRequestRequest) RawResponse(v string) *ContinueInterceptedRequestRequest {
	r.opts["rawResponse"] = v
	return r
}

// If set the request url will be modified in a way that's not observable by page. (optional)
func (r *ContinueInterceptedRequestRequest) URL(v string) *ContinueInterceptedRequestRequest {
	r.opts["url"] = v
	return r
}

// If set this allows the request method to be overridden. (optional)
func (r *ContinueInterceptedRequestRequest) Method(v string) *ContinueInterceptedRequestRequest {
	r.opts["method"] = v
	return r
}

// If set this allows postData to be set. (optional)
func (r *ContinueInterceptedRequestRequest) PostData(v string) *ContinueInterceptedRequestRequest {
	r.opts["postData"] = v
	return r
}

// If set this allows the request headers to be changed. (optional)
func (r *ContinueInterceptedRequestRequest) Headers(v *Headers) *ContinueInterceptedRequestRequest {
	r.opts["headers"] = v
	return r
}

func (r *ContinueInterceptedRequestRequest) Do() error {
	return r.client.Call("Network.continueInterceptedRequest", r.opts, nil)
}

func init() {
	rpc.EventTypes["Network.resourceChangedPriority"] = func() interface{} { return new(ResourceChangedPriorityEvent) }
	rpc.EventTypes["Network.requestWillBeSent"] = func() interface{} { return new(RequestWillBeSentEvent) }
	rpc.EventTypes["Network.requestServedFromCache"] = func() interface{} { return new(RequestServedFromCacheEvent) }
	rpc.EventTypes["Network.responseReceived"] = func() interface{} { return new(ResponseReceivedEvent) }
	rpc.EventTypes["Network.dataReceived"] = func() interface{} { return new(DataReceivedEvent) }
	rpc.EventTypes["Network.loadingFinished"] = func() interface{} { return new(LoadingFinishedEvent) }
	rpc.EventTypes["Network.loadingFailed"] = func() interface{} { return new(LoadingFailedEvent) }
	rpc.EventTypes["Network.webSocketWillSendHandshakeRequest"] = func() interface{} { return new(WebSocketWillSendHandshakeRequestEvent) }
	rpc.EventTypes["Network.webSocketHandshakeResponseReceived"] = func() interface{} { return new(WebSocketHandshakeResponseReceivedEvent) }
	rpc.EventTypes["Network.webSocketCreated"] = func() interface{} { return new(WebSocketCreatedEvent) }
	rpc.EventTypes["Network.webSocketClosed"] = func() interface{} { return new(WebSocketClosedEvent) }
	rpc.EventTypes["Network.webSocketFrameReceived"] = func() interface{} { return new(WebSocketFrameReceivedEvent) }
	rpc.EventTypes["Network.webSocketFrameError"] = func() interface{} { return new(WebSocketFrameErrorEvent) }
	rpc.EventTypes["Network.webSocketFrameSent"] = func() interface{} { return new(WebSocketFrameSentEvent) }
	rpc.EventTypes["Network.eventSourceMessageReceived"] = func() interface{} { return new(EventSourceMessageReceivedEvent) }
	rpc.EventTypes["Network.requestIntercepted"] = func() interface{} { return new(RequestInterceptedEvent) }
}

// Fired when resource loading priority is changed (experimental)
type ResourceChangedPriorityEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// New priority
	NewPriority ResourcePriority `json:"newPriority"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`
}

// Fired when page is about to send HTTP request.
type RequestWillBeSentEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Frame identifier.
	FrameId string `json:"frameId"`

	// Loader identifier.
	LoaderId LoaderId `json:"loaderId"`

	// URL of the document this request is loaded for.
	DocumentURL string `json:"documentURL"`

	// Request data.
	Request *Request `json:"request"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// UTC Timestamp.
	WallTime Timestamp `json:"wallTime"`

	// Request initiator.
	Initiator *Initiator `json:"initiator"`

	// Redirect response data. (optional)
	RedirectResponse *Response `json:"redirectResponse"`

	// Type of this resource. (optional, experimental)
	Type string `json:"type"`
}

// Fired if request ended up loading from cache.
type RequestServedFromCacheEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`
}

// Fired when HTTP response is available.
type ResponseReceivedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Frame identifier.
	FrameId string `json:"frameId"`

	// Loader identifier.
	LoaderId LoaderId `json:"loaderId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// Resource type.
	Type string `json:"type"`

	// Response data.
	Response *Response `json:"response"`
}

// Fired when data chunk was received over the network.
type DataReceivedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// Data chunk length.
	DataLength int `json:"dataLength"`

	// Actual bytes received (might be less than dataLength for compressed encodings).
	EncodedDataLength int `json:"encodedDataLength"`
}

// Fired when HTTP request has finished loading.
type LoadingFinishedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// Total number of bytes received for this request.
	EncodedDataLength float64 `json:"encodedDataLength"`
}

// Fired when HTTP request has failed to load.
type LoadingFailedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// Resource type.
	Type string `json:"type"`

	// User friendly error message.
	ErrorText string `json:"errorText"`

	// True if loading was canceled. (optional)
	Canceled bool `json:"canceled"`

	// The reason why loading was blocked, if any. (optional, experimental)
	BlockedReason BlockedReason `json:"blockedReason"`
}

// Fired when WebSocket is about to initiate handshake. (experimental)
type WebSocketWillSendHandshakeRequestEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// UTC Timestamp.
	WallTime Timestamp `json:"wallTime"`

	// WebSocket request data.
	Request *WebSocketRequest `json:"request"`
}

// Fired when WebSocket handshake response becomes available. (experimental)
type WebSocketHandshakeResponseReceivedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// WebSocket response data.
	Response *WebSocketResponse `json:"response"`
}

// Fired upon WebSocket creation. (experimental)
type WebSocketCreatedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// WebSocket request URL.
	URL string `json:"url"`

	// Request initiator. (optional)
	Initiator *Initiator `json:"initiator"`
}

// Fired when WebSocket is closed. (experimental)
type WebSocketClosedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`
}

// Fired when WebSocket frame is received. (experimental)
type WebSocketFrameReceivedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// WebSocket response data.
	Response *WebSocketFrame `json:"response"`
}

// Fired when WebSocket frame error occurs. (experimental)
type WebSocketFrameErrorEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// WebSocket frame error message.
	ErrorMessage string `json:"errorMessage"`
}

// Fired when WebSocket frame is sent. (experimental)
type WebSocketFrameSentEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// WebSocket response data.
	Response *WebSocketFrame `json:"response"`
}

// Fired when EventSource message is received. (experimental)
type EventSourceMessageReceivedEvent struct {
	// Request identifier.
	RequestId RequestId `json:"requestId"`

	// Timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// Message type.
	EventName string `json:"eventName"`

	// Message identifier.
	EventId string `json:"eventId"`

	// Message content.
	Data string `json:"data"`
}

// Details of an intercepted HTTP request, which must be either allowed, blocked, modified or mocked. (experimental)
type RequestInterceptedEvent struct {
	// Each request the page makes will have a unique id, however if any redirects are encountered while processing that fetch, they will be reported with the same id as the original fetch.
	InterceptionId InterceptionId `json:"InterceptionId"`

	Request *Request `json:"request"`

	// How the requested resource will be used.
	ResourceType string `json:"resourceType"`

	// HTTP response headers, only sent if a redirect was intercepted. (optional)
	RedirectHeaders *Headers `json:"redirectHeaders"`

	// HTTP response code, only sent if a redirect was intercepted. (optional)
	RedirectStatusCode int `json:"redirectStatusCode"`

	// Redirect location, only sent if a redirect was intercepted. (optional)
	RedirectUrl string `json:"redirectUrl"`
}

package httptrace

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

var (
	// RedactedHeaders is a slice of header names whose values should be
	// entirely redacted from logs.
	RedactedHeaders = []string{"Authorization"}
)

func init() { appdash.RegisterEvent(ClientEvent{}) }

// NewClientEvent returns an event which records various aspects of an
// HTTP request.  The returned value is incomplete, and should have
// the response status, size, and the ClientSend/ClientRecv times set
// before being logged.
func NewClientEvent(r *http.Request) *ClientEvent {
	return &ClientEvent{Request: requestInfo(r)}
}

// RequestInfo describes an HTTP request.
type RequestInfo struct {
	Method        string
	URI           string
	Proto         string
	Headers       map[string]string
	Host          string
	RemoteAddr    string
	ContentLength int64
}

func requestInfo(r *http.Request) RequestInfo {
	return RequestInfo{
		Method:        r.Method,
		URI:           r.URL.RequestURI(),
		Proto:         r.Proto,
		Headers:       redactHeaders(r.Header, r.Trailer),
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
		ContentLength: r.ContentLength,
	}
}

// ClientEvent records an HTTP client request event.
type ClientEvent struct {
	Request    RequestInfo  `trace:"Client.Request"`
	Response   ResponseInfo `trace:"Client.Response"`
	ClientSend time.Time    `trace:"Client.Send"`
	ClientRecv time.Time    `trace:"Client.Recv"`
}

// Schema returns the constant "HTTPClient".
func (ClientEvent) Schema() string { return "HTTPClient" }

// Important implements the appdash ImportantEvent.
func (ClientEvent) Important() []string {
	return []string{
		"Client.Request.Headers.If-Modified-Since",
		"Client.Request.Headers.If-None-Match",
		"Client.Response.StatusCode",
	}
}

// Start implements the appdash TimespanEvent interface.
func (e ClientEvent) Start() time.Time { return e.ClientSend }

// End implements the appdash TimespanEvent interface.
func (e ClientEvent) End() time.Time { return e.ClientRecv }

var (
	redacted = []string{"REDACTED"}
)

func redactHeaders(header, trailer http.Header) map[string]string {
	h := make(http.Header, len(header)+len(trailer))
	for k, v := range header {
		if isRedacted(k) {
			h[k] = redacted
		} else {
			h[k] = v
		}
	}
	for k, v := range trailer {
		if isRedacted(k) {
			h[k] = redacted
		} else {
			h[k] = append(h[k], v...)
		}
	}
	m := make(map[string]string, len(h))
	for k, v := range h {
		m[http.CanonicalHeaderKey(k)] = strings.Join(v, ",")
	}
	return m
}

func isRedacted(name string) bool {
	for _, v := range RedactedHeaders {
		if strings.EqualFold(name, v) {
			return true
		}
	}
	return false
}

// Transport is an HTTP transport that adds appdash span ID headers
// to requests so that downstream operations are associated with the
// same trace.
type Transport struct {
	// Recorder is the current span's recorder. A new child Recorder
	// (with a new child SpanID) is created for each HTTP roundtrip.
	*appdash.Recorder

	// Transport is the underlying HTTP transport to use when making
	// requests.  It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper

	SetName bool

	// requests keeps clone request
	reqMu    sync.Mutex
	requests map[*http.Request]*http.Request
}

// RoundTrip implements the RoundTripper interface.
func (t *Transport) RoundTrip(original *http.Request) (*http.Response, error) {
	// To set extra querystring params, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	req := cloneRequest(original)
	t.setCloneRequest(original, req)
	defer t.setCloneRequest(original, nil)

	child := t.Recorder.Child()
	if t.SetName {
		child.Name("Request " + req.URL.Host)
	}

	// New child span is created and set as HTTP header instead of using `child`
	// in order to have a single span recording operation per httptrace event
	// (HTTPClient or HTTPServer).
	span := appdash.NewSpanID(t.Recorder.SpanID)

	SetSpanIDHeader(req.Header, span)

	e := NewClientEvent(req)
	e.ClientSend = time.Now()

	// Make the HTTP request.
	transport := t.getTransport()
	resp, err := transport.RoundTrip(req)

	e.ClientRecv = time.Now()
	if err == nil {
		e.Response = responseInfo(resp)
	} else {
		e.Response.StatusCode = -1
	}
	child.Event(e)
	child.Finish()
	return resp, err
}

// cloneRequest returns a clone of the provided *http.Request. The clone is a
// shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}
	return r2
}

// setCloneRequest keeps track of the cloned based on the original request so
// that it can be canceled in the future in CancelRequest.
func (t *Transport) setCloneRequest(original *http.Request, clone *http.Request) {
	t.reqMu.Lock()
	defer t.reqMu.Unlock()
	if t.requests == nil {
		t.requests = make(map[*http.Request]*http.Request)
	}

	if clone != nil {
		t.requests[original] = clone
	} else {
		delete(t.requests, original)
	}
}

// getTransport returns custom transport or DefaultTransport
func (t *Transport) getTransport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// CancelRequest cancels an in-flight request by closing its connection.
func (t *Transport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}

	newReq, ok := t.requests[req]
	if !ok {
		return
	}

	transport := t.getTransport()
	if ts, ok := transport.(canceler); ok {
		ts.CancelRequest(newReq)
	}
}

package httptrace

import (
	"log"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

func init() { appdash.RegisterEvent(ServerEvent{}) }

// NewServerEvent returns an event which records various aspects of an
// HTTP response. It takes an HTTP request, not response, as input
// because the information it records is derived from the request, and
// HTTP handlers don't have access to the response struct (only
// http.ResponseWriter, which requires wrapping or buffering to
// introspect).
//
// The returned value is incomplete and should have its Response and
// ServerRecv/ServerSend values set before being logged.
func NewServerEvent(r *http.Request) *ServerEvent {
	return &ServerEvent{Request: requestInfo(r)}
}

// ResponseInfo describes an HTTP response.
type ResponseInfo struct {
	Headers       map[string]string
	ContentLength int64
	StatusCode    int
}

func responseInfo(r *http.Response) ResponseInfo {
	return ResponseInfo{
		Headers:       redactHeaders(r.Header, r.Trailer),
		ContentLength: r.ContentLength,
		StatusCode:    r.StatusCode,
	}
}

// ServerEvent records an HTTP server request handling event.
type ServerEvent struct {
	Request    RequestInfo  `trace:"Server.Request"`
	Response   ResponseInfo `trace:"Server.Response"`
	Route      string       `trace:"Server.Route"`
	User       string       `trace:"Server.User"`
	ServerRecv time.Time    `trace:"Server.Recv"`
	ServerSend time.Time    `trace:"Server.Send"`
}

// Schema returns the constant "HTTPServer".
func (ServerEvent) Schema() string { return "HTTPServer" }

// Important implements the appdash ImportantEvent.
func (ServerEvent) Important() []string {
	return []string{"Server.Response.StatusCode"}
}

// Start implements the appdash TimespanEvent interface.
func (e ServerEvent) Start() time.Time { return e.ServerRecv }

// End implements the appdash TimespanEvent interface.
func (e ServerEvent) End() time.Time { return e.ServerSend }

// Middleware creates a new http.Handler middleware
// (negroni-compliant) that records incoming HTTP requests to the
// collector c as "HTTPServer"-schema events.
func Middleware(c appdash.Collector, conf *MiddlewareConfig) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		spanID, spanFromHeader, err := getSpanID(r.Header)
		if err != nil {
			log.Printf("Warning: invalid %s header: %s. (Continuing with request handling.)", spanFromHeader, err)
		}
		usingProvidedSpanID := (spanFromHeader == HeaderSpanID)

		if conf.SetContextSpan != nil {
			conf.SetContextSpan(r, *spanID)
		}

		e := NewServerEvent(r)
		e.ServerRecv = time.Now()

		rr := &responseInfoRecorder{ResponseWriter: rw}
		next(rr, r)
		SetSpanIDHeader(rr.Header(), *spanID)

		if !usingProvidedSpanID {
			e.Request = requestInfo(r)
		}
		if conf.RouteName != nil {
			e.Route = conf.RouteName(r)
		}
		if conf.CurrentUser != nil {
			e.User = conf.CurrentUser(r)
		}
		e.Response = responseInfo(rr.partialResponse())
		e.ServerSend = time.Now()

		rec := appdash.NewRecorder(*spanID, c)
		if e.Route != "" {
			rec.Name("Serve " + e.Route)
		} else {
			rec.Name("Serve " + r.URL.Host + r.URL.Path)
		}
		rec.Event(e)
		rec.Finish()
	}
}

// MiddlewareConfig configures the HTTP tracing middleware.
type MiddlewareConfig struct {
	// RouteName, if non-nil, is called to get the current route's
	// name. This name is used as the span's name.
	RouteName func(*http.Request) string

	// CurrentUser, if non-nil, is called to get the current user ID
	// (which may be a login or a numeric ID).
	CurrentUser func(*http.Request) string

	// SetContextSpan, if non-nil, is called to set the span (which is
	// either taken from the client request header or created anew) in
	// the HTTP request context, so it may be used by other parts of
	// the handling process.
	SetContextSpan func(*http.Request, appdash.SpanID)
}

// responseInfoRecorder is an http.ResponseWriter that records a
// response's HTTP status code and body length and forwards all
// operations onto an underlying http.ResponseWriter, without
// buffering the response body.
type responseInfoRecorder struct {
	statusCode    int   // HTTP response status code
	ContentLength int64 // number of bytes written using the Write method

	http.ResponseWriter // underlying ResponseWriter to pass-thru to
}

// Write always succeeds and writes to r.Body, if not nil.
func (r *responseInfoRecorder) Write(b []byte) (int, error) {
	r.ContentLength += int64(len(b))
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func (r *responseInfoRecorder) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

// WriteHeader sets r.Code.
func (r *responseInfoRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// partialResponse constructs a partial response object based on the
// information it is able to determine about the response.
func (r *responseInfoRecorder) partialResponse() *http.Response {
	return &http.Response{
		StatusCode:    r.StatusCode(),
		ContentLength: r.ContentLength,
		Header:        r.Header(),
	}
}

// Flush implements the http.Flusher interface and sends any buffered
// data to the client, if the underlying http.ResponseWriter itself
// implements http.Flusher.
func (r *responseInfoRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

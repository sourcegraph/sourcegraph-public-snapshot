package httpwrapper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
)

const kJoinIdsKey = "X-Traceguide-Join-Ids"
const kClientSpanGuidKey = "X-Traceguide-Client-Span-Guid"
const kCookieKey = "traceguide_session_id"

type ServerConfig struct {
	// Called on each request with the span created for that request.
	// Should call ActiveSpan.SetOperation.
	//
	// For example, with Gorilla mux:
	//   WithActiveSpanFunc: func (req *http.Request, s instrument.ActiveSpan) {
	//     s.SetOperation("gorilla/mux/" + mux.CurrentRoute(req).GetName())
	//   }
	WithActiveSpanFunc func(*http.Request, instrument.ActiveSpan)

	// TODO other configuration, e.g. should we log the body
}

func maybeParseAndLogPayload(eventName string, s instrument.ActiveSpan, body []byte) {
	jsonBody := make(map[string]interface{})
	// TODO do this attempted parsing during backend ingestion
	err := json.Unmarshal(body, &jsonBody)
	if err == nil {
		s.Log(instrument.EventName(eventName).Payload(jsonBody))
	} else {
		s.Log(instrument.EventName(eventName).Payload(string(body)))
	}
}

// Calls next with rw and r inside of a Span.
//
// Join IDs may be added using config.WithActiveSpan. In addition,
// Join IDs are populated automatically using the following
// mechanisms:
//
//   - cookie: The cookie "traceguide_session_id" if present will be
//     used to set a Join Id whose key is "traceguide_session_id" and
//     whose value is the value of the cookie.  This is most
//     appropriate for single user Runtimes (e.g. browsers),
//     especially where some other Runtime will be able to associate
//     an end user this session.
//
//   - headers: The header "X-Traceguide-Join-Ids" if present will be
//     used to define Join Ids of the span associated with a request.
//     Each value of the header should be of the form "key=value".  In
//     addition, the value of the header
//     "X-Traceguide-Client-Span-Guid" is used to create the parent
//	   span association.
//
func MakeMiddleware(config *ServerConfig) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		f := func(s instrument.ActiveSpan) error {
			// Copy Traceguide-specific data into the span or a log record.
			for _, id := range req.Header[kJoinIdsKey] {
				ss := strings.SplitN(id, "=", 2) // assume first '=' is the key/value split
				s.AddTraceJoinId(ss[0], ss[1])
			}
			if cookie, err := req.Cookie(kCookieKey); err == nil {
				// This key doesn't need to make the cookie one, but why not?
				s.AddTraceJoinId(kCookieKey, cookie.Value)
			}
			config.WithActiveSpanFunc(req, s)
			if guid, found := req.Header[kClientSpanGuidKey]; found && len(guid) == 1 {
				s.AddAttribute("parent_span_guid", guid[0])
			}

			// Duplicate the request and read the body
			body, err := ioutil.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				http.Error(rw, "Unable to read request", http.StatusInternalServerError)
				return err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			maybeParseAndLogPayload("HTTP request body", s, body)
			rec := responseRecorder{ResponseWriter: rw}

			// Call the underlying handler.
			next(&rec, req)

			if rec.StatusCode() != http.StatusOK {
				s.Log(instrument.FileLine(1).Error().
					Printf("HTTP status code: %d", rec.StatusCode()))
			}
			maybeParseAndLogPayload("HTTP response body", s, rec.buf.Bytes())

			// TODO write new join ads into response header?
			return nil
		}
		_ = instrument.RunInSpan(f, instrument.OnStack)
	}
}

// Wraps the given handler.
func Handler(operation string, handler http.HandlerFunc) http.HandlerFunc {
	m := MakeMiddleware(&ServerConfig{
		WithActiveSpanFunc: func(_ *http.Request, s instrument.ActiveSpan) {
			s.SetOperation(operation)
		},
	})
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		m(rw, req, handler)
	})
}

// Code below inspired by sqs@sourcegraph.com

type responseRecorder struct {
	statusCode int
	buf        bytes.Buffer

	http.ResponseWriter
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	r.buf.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

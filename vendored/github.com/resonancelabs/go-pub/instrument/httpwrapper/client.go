package httpwrapper

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
)

type ClientConfig struct {
	// Called on each request with the span created for that request.
	// Should call ActiveSpan.SetOperation.
	WithActiveSpanFunc func(req *http.Request, s instrument.ActiveSpan)

	// TODO other options such as what to log
}

type Transport struct {
	http.Transport
	Config *ClientConfig
}

// NewClient returns a new *http.Client that uses a wrapper transport.
// For example:
//
//	client := httpwrapper.NewClient(
//		&httpwrapper.ClientConfig{
//			WithActiveSpanFunc: func(req *http.Request, s instrument.ActiveSpan) {
//				s.SetOperation(op)
//			},
//		})
//	resp, err := client.Post(url, "text/plain", body))
func NewClient(config *ClientConfig) *http.Client {
	return &http.Client{
		Transport: NewTransport(config),
	}
}

// NewTransport creates a new Transport that wraps each request in a
// Span. See also NewClient.
//
// For example:
//
//	client := &http.Client{
//		Transport: httpwrapper.NewTransport(
//			&httpwrapper.ClientConfig{
//				WithActiveSpanFunc: func(req *http.Request, s instrument.ActiveSpan) {
//					s.SetOperation(op)
//				},
//			})}
//	resp, err := client.Post(url, "text/plain", body))
//
func NewTransport(config *ClientConfig) *Transport {
	return &Transport{
		Config: config,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	f := func(s instrument.ActiveSpan) error {
		t.Config.WithActiveSpanFunc(req, s)
		s.MergeTraceJoinIdsFromStack()
		s.Log(instrument.EventName("HTTP request URL").Payload(req.URL.String()))

		// RoundTrippers should not modify requests, so make a copy of
		// this one.
		dupReq := *req
		dupReq.Header = make(http.Header)
		for k, v := range req.Header {
			dupReq.Header[k] = v
		}
		// Add join ids into a header
		for k, v := range s.TraceJoinIds() {
			dupReq.Header.Add(kJoinIdsKey, fmt.Sprintf("%s=%s", k, v))
		}
		// Add the span id as well
		dupReq.Header.Add(kClientSpanGuidKey, string(s.Guid()))

		// Make the HTTP request.
		resp, err = t.Transport.RoundTrip(&dupReq)

		// TODO: could record response
		return err
	}
	err = instrument.RunInSpan(f, instrument.OnStack)
	return resp, err
}

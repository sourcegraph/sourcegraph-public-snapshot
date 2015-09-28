package httputil

import (
	"io"
	"net/http"
	"net/http/httputil"
)

// A TracingTransport prints out the full HTTP request and response
// for each roundtrip.
type TracingTransport struct {
	io.Writer                   // destination of trace output
	Transport http.RoundTripper // underlying transport (or default if nil)
}

func (t *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var u http.RoundTripper
	if t.Transport != nil {
		u = t.Transport
	} else {
		u = http.DefaultTransport
	}

	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	t.Writer.Write(reqBytes)

	resp, err := u.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respBytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	t.Writer.Write(respBytes)

	return resp, nil
}

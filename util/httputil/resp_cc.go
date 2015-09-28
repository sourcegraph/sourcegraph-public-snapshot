package httputil

import "net/http"

type ModifyResponseTransport struct {
	Transport      http.RoundTripper
	ModifyResponse func(*http.Request, *http.Response)
}

func (t *ModifyResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	resp, err := transport.RoundTrip(req)
	if resp != nil {
		t.ModifyResponse(req, resp)
	}
	return resp, err
}

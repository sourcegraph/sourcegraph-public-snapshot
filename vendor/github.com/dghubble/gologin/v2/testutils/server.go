package testutils

import (
	"net/http"
	"net/http/httptest"
	"net/url"
)

// TestServer returns a new httptest.Server, its ServeMux for adding handlers,
// and a client which proxies requests to the server using a custom transport.
// The caller must close the server.
func TestServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	transport := &RewriteTransport{&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}}
	client := &http.Client{Transport: transport}
	return client, mux, server
}

// NewErrorServer returns a new httptest.Server, which responds with the given
// error message and code, and a client which proxies requests to the server
// using a custom transport. The caller must close the server.
func NewErrorServer(message string, code int) (*http.Client, *httptest.Server) {
	client, mux, server := TestServer()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, message, code)
	})
	return client, server
}

// NewTestServerFunc is an adapter to allow the use of ordinary functions as
// httptest.Server's for testing. Caller must close the server.
func NewTestServerFunc(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

// RewriteTransport rewrites https requests to http to avoid TLS cert issues
// during testing.
type RewriteTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

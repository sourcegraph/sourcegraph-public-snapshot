package headertransport

import "net/http"

// AddHeaderTransport implements http.RoundTripper and allows us to inject
// additional HTTP headers on requests
type AddHeaderTransport struct {
	T http.RoundTripper

	additionalHeaders map[string]string
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for header, value := range adt.additionalHeaders {
		req.Header.Add(header, value)
	}
	return adt.T.RoundTrip(req)
}

func New(t http.RoundTripper, headers map[string]string) *AddHeaderTransport {
	if t == nil {
		t = http.DefaultTransport
	}
	return &AddHeaderTransport{
		T:                 t,
		additionalHeaders: headers,
	}
}

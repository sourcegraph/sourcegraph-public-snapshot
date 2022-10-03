package monitoring

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

func NewAddHeaderTransport(T http.RoundTripper, headers map[string]string) *AddHeaderTransport {
	if T == nil {
		T = http.DefaultTransport
	}
	return &AddHeaderTransport{
		T:                 T,
		additionalHeaders: headers,
	}
}

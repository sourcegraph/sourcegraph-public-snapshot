package s3

import (
	"net/http"
	"time"

	"github.com/kr/http/transport"
)

// Client returns an HTTP client that signs all outgoing requests.
// The returned Transport also provides CancelRequest.
// Client is equivalent to DefaultService.Client.
func Client(k Keys) *http.Client {
	return DefaultService.Client(k)
}

// Client returns an HTTP client that signs all outgoing requests.
// The returned Transport also provides CancelRequest.
func (s *Service) Client(k Keys) *http.Client {
	tr := &transport.Wrapper{Modify: func(r *http.Request) error {
		if r.Header.Get("Date") == "" {
			r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
		}
		s.Sign(r, k)
		return nil
	}}
	return &http.Client{Transport: tr}
}

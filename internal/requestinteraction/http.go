package requestinteraction

import (
	"net/http"
)

const (
	// Sourcegraph-specific header key for propagating an interaction ID.
	headerKeyInteractionID = "X-Sourcegraph-Interaction-ID"
)

type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	interaction := FromContext(req.Context())
	if interaction != nil {
		req.Header.Set(headerKeyInteractionID, interaction.ID)
	}

	return t.RoundTripper.RoundTrip(req)
}

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctxWithClient := WithClient(req.Context(), &Interaction{
			ID: req.Header.Get(headerKeyInteractionID),
		})
		next.ServeHTTP(rw, req.WithContext(ctxWithClient))
	})
}

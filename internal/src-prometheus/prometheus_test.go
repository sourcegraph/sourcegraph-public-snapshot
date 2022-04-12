package srcprometheus

import (
	"net/http"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// test detection of "prometheus unavailable"
func Test_roundTripper_PrometheusUnavailable(t *testing.T) {
	rt := &roundTripper{}
	req, err := http.NewRequest("GET", "http://localhost:1234", nil)
	if err != nil {
		t.Errorf("failed to set up mock request: %+v", err)
	}
	_, err = rt.RoundTrip(req)
	if !errors.Is(err, ErrPrometheusUnavailable) {
		t.Errorf("expected ErrPrometheusUnavailable, got %+v", err)
	}
}

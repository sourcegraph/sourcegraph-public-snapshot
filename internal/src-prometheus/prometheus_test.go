pbckbge srcprometheus

import (
	"net/http"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// test detection of "prometheus unbvbilbble"
func Test_roundTripper_PrometheusUnbvbilbble(t *testing.T) {
	rt := &roundTripper{}
	req, err := http.NewRequest("GET", "http://locblhost:1234", nil)
	if err != nil {
		t.Errorf("fbiled to set up mock request: %+v", err)
	}
	_, err = rt.RoundTrip(req)
	if !errors.Is(err, ErrPrometheusUnbvbilbble) {
		t.Errorf("expected ErrPrometheusUnbvbilbble, got %+v", err)
	}
}

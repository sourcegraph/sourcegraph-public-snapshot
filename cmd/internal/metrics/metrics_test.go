package metrics

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}

func doRequest(hc *http.Client, u string) error {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	_, err = hc.Do(req)
	return err
}

func TestRequestMeterTransport(t *testing.T) {
	rm := NewRequestMeter("foosystem", "Total number of requests sent to foosystem.")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 10)
		_, err := w.Write([]byte("the quick brown fox jumps over the lazy dog"))
		if err != nil {
			t.Error(err)
		}
	})
	hc, teardown := testingHTTPClient(h)
	defer teardown()

	hc.Transport = rm.Transport(hc.Transport, func(u *url.URL) string {
		return u.Path
	})

	err := doRequest(hc, "http://foosystem.com/apiCallA")
	if err != nil {
		t.Error(err)
	}

	err = doRequest(hc, "http://foosystem.com/apiCallB")
	if err != nil {
		t.Error(err)
	}

	c, err := rm.counter.GetMetricWith(map[string]string{"category": "/apiCallA", "code": "200"})
	if err != nil {
		t.Error(err)
	}
	val := testutil.ToFloat64(c)

	if val != 1.0 {
		t.Errorf("expected counter == 1, got %f", val)
	}
}

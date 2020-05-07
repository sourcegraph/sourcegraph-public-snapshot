package metrics

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
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

	c, err := rm.counter.GetMetricWith(map[string]string{"category": "/apiCallA", "code": "200", "host": "foosystem.com"})
	if err != nil {
		t.Error(err)
	}
	val := testutil.ToFloat64(c)

	if val != 1.0 {
		t.Errorf("expected counter == 1, got %f", val)
	}
}

func TestMustRegisterDiskMonitor(t *testing.T) {
	registry := prometheus.NewPedanticRegistry()
	registerer = registry
	defer func() { registerer = prometheus.DefaultRegisterer }()

	want := []string{}
	for i := 0; i <= 2; i++ {
		path, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			path := path
			_ = os.RemoveAll(path)
		})
		// Register twice to ensure we don't panic and we don't collect twice.
		MustRegisterDiskMonitor(path)
		MustRegisterDiskMonitor(path)
		want = append(want,
			fmt.Sprintf("src_disk_space_available_bytes{path=%s}", path),
			fmt.Sprintf("src_disk_space_total_bytes{path=%s}", path))
	}

	mfs, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, mf := range mfs {
		for _, m := range mf.Metric {
			var labels []string
			for _, l := range m.Label {
				labels = append(labels, fmt.Sprintf("%s=%s", *l.Name, *l.Value))
			}
			got = append(got, fmt.Sprintf("%s{%s}", *mf.Name, strings.Join(labels, " ")))
		}
	}

	sort.Strings(want)
	sort.Strings(got)
	if !cmp.Equal(want, got) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestREDClientRecord(t *testing.T) {
	rm := NewREDClient("foosystems")
	rec := rm.Record("foo")
	func(err error) {
		rec(err)
	}(nil)

	val := testutil.ToFloat64(rm.cli.Count)
	if got, want := val, 1.0; got != want {
		t.Errorf("expected entries in histogram, got %f, expected %f", got, want)
	}

	m, err := rm.cli.Count.GetMetricWith(prometheus.Labels{"method": "foo"})
	if err != nil {
		t.Fatal(err)
	}
	val = testutil.ToFloat64(m)
	if got, want := val, 1.0; got != want {
		t.Errorf("expected counter to inc, got %f, expected %f", got, want)
	}

	m, err = rm.cli.Errors.GetMetricWith(prometheus.Labels{"method": "foo"})
	if err != nil {
		t.Fatal(err)
	}
	val = testutil.ToFloat64(m)
	if got, want := val, 0.0; got != want {
		t.Errorf("error counter should not increment, got %f, expected %f", got, want)
	}
}

func TestREDClientRecord_Error(t *testing.T) {
	rm := NewREDClient("failure")
	rec := rm.Record("foo")
	func(err error) {
		rec(err)
	}(errors.New("boom"))

	val := testutil.ToFloat64(rm.cli.Count)
	if got, want := val, 1.0; got != want {
		t.Errorf("expected entries in histogram, got %f, expected %f", got, want)
	}

	m, err := rm.cli.Count.GetMetricWith(prometheus.Labels{"method": "foo"})
	if err != nil {
		t.Fatal(err)
	}
	val = testutil.ToFloat64(m)
	if got, want := val, 1.0; got != want {
		t.Errorf("expected counter to inc, got %f, expected %f", got, want)
	}

	m, err = rm.cli.Errors.GetMetricWith(prometheus.Labels{"method": "foo"})
	if err != nil {
		t.Fatal(err)
	}
	val = testutil.ToFloat64(m)
	if got, want := val, 1.0; got != want {
		t.Errorf("error counter should increment, got %f, expected %f", got, want)
	}
}

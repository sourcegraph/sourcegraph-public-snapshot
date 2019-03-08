package metrics

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// RequestCounter wraps a Prometheus request counter that is incremented by requests made by derived
// http.RoundTrippers.
type RequestCounter struct {
	counter   *prometheus.CounterVec
	subsystem string
}

// NewRequestCounter creates a new request counter.
func NewRequestCounter(subsystem, help string) *RequestCounter {
	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      help,
	}, []string{"category", "code"})
	prometheus.MustRegister(requestCounter)
	return &RequestCounter{counter: requestCounter, subsystem: subsystem}
}

// Transport returns an http.RoundTripper that increments c for each request. The categoryFunc is called to
// determine the category label for each request.
func (c *RequestCounter) Transport(transport http.RoundTripper, categoryFunc func(*url.URL) string) http.RoundTripper {
	return &requestCounterMiddleware{
		counter:      c,
		transport:    transport,
		categoryFunc: categoryFunc,
	}
}

// Doer returns an httpcli.Doer that increments c for each request. The categoryFunc is called to
// determine the category label for each request.
func (c *RequestCounter) Doer(cli httpcli.Doer, categoryFunc func(*url.URL) string) httpcli.Doer {
	return &requestCounterMiddleware{
		counter:      c,
		cli:          cli,
		categoryFunc: categoryFunc,
	}
}

type requestCounterMiddleware struct {
	counter      *RequestCounter
	cli          httpcli.Doer
	transport    http.RoundTripper
	categoryFunc func(*url.URL) string
}

func (t *requestCounterMiddleware) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	if t.transport != nil {
		resp, err = t.transport.RoundTrip(r)
	} else if t.cli != nil {
		resp, err = t.cli.Do(r)
	}

	category := t.categoryFunc(r.URL)

	var code string
	if err != nil {
		code = "error"
	} else {
		code = strconv.Itoa(resp.StatusCode)
	}

	t.counter.counter.WithLabelValues(category, code).Inc()
	log15.Debug("TRACE "+t.counter.subsystem, "host", r.URL.Host, "path", r.URL.Path, "code", code, "duration", time.Since(start))
	return
}

func (t *requestCounterMiddleware) Do(req *http.Request) (*http.Response, error) {
	return t.RoundTrip(req)
}

package githubutil

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/prometheus/client_golang/prometheus"
)

var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "github",
	Name:      "requests_total",
	Help:      "Total number of requests sent to the GitHub API.",
}, []string{"category", "code"})

func init() {
	prometheus.MustRegister(requestCount)
}

// metricsTransport wraps a transport for the GitHub API to export metrics to
// prometheus
type metricsTransport struct {
	Transport http.RoundTripper
}

func (t *metricsTransport) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	resp, err = t.Transport.RoundTrip(r)

	// The first component of the Path mostly maps to the type of API
	// request we are making. See `curl https://api.github.com` for the
	// exact mapping
	var category string
	if parts := strings.SplitN(r.URL.Path, "/", 3); len(parts) > 1 {
		category = parts[1]
	}

	var code string
	if err != nil {
		code = "error"
	} else {
		code = strconv.Itoa(resp.StatusCode)
	}

	requestCount.WithLabelValues(category, code).Inc()
	log15.Debug("TRACE github", "path", r.URL.Path, "code", code, "duration", time.Since(start))
	return
}

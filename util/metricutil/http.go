package metricutil

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

var metricLabels = []string{"route", "method", "code"}
var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "requests_total",
	Help:      "Total number of HTTP requests made.",
}, metricLabels)
var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "request_duration_seconds",
	Help:      "The HTTP request latencies in seconds.",
	Buckets:   []float64{1, 5, 10, 60, 300},
}, metricLabels)
var requestHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "requests_last_timestamp_unixtime",
	Help:      "Last time a request finished for a http endpoint.",
}, metricLabels)

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestHeartbeat)
}

func HTTPMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	rwIntercept := &ResponseWriterStatusIntercept{ResponseWriter: rw}
	next(rwIntercept, r)

	// If we have an error, name is an empty string which
	// indicates to httptrace to use a fallback value
	name, _ := httpctx.RouteNameOrError(r)
	duration := time.Now().Sub(start)
	labels := prometheus.Labels{
		"route":  name,
		"method": strings.ToLower(r.Method),
		"code":   strconv.Itoa(rwIntercept.Code),
	}
	requestCount.With(labels).Inc()
	requestDuration.With(labels).Observe(duration.Seconds())
	requestHeartbeat.With(labels).Set(float64(time.Now().Unix()))

	log15.Debug("HTTP Request", "method", r.Method, "URL", r.URL.String(), "routename", name, "duration", duration, "code", rwIntercept.Code)
}

// ResponseWriterStatusIntercept implements the http.ResponseWriter interface
// so we can intercept the status that we can otherwise not access
type ResponseWriterStatusIntercept struct {
	http.ResponseWriter
	Code int
}

// WriteHeader saves the code and then delegates to http.ResponseWriter
func (r *ResponseWriterStatusIntercept) WriteHeader(code int) {
	r.Code = code
	r.ResponseWriter.WriteHeader(code)
}

var _ http.ResponseWriter = (*ResponseWriterStatusIntercept)(nil)

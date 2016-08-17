package httptrace

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

type key int

const (
	routeNameKey key = iota
)

var metricLabels = []string{"route", "method", "code", "repo"}
var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "request_duration_seconds",
	Help:      "The HTTP request latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, metricLabels)
var requestHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "requests_last_timestamp_unixtime",
	Help:      "Last time a request finished for a http endpoint.",
}, metricLabels)

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestHeartbeat)
}

// Middleware captures and exports metrics to Prometheus, etc.
func Middleware(next http.Handler, sessionInfo func(*http.Request) (uid, sessionID string)) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		routeName := "unknown"
		r = r.WithContext(context.WithValue(r.Context(), routeNameKey, &routeName))

		rwIntercept := &ResponseWriterStatusIntercept{ResponseWriter: rw}
		if traceutil.DefaultCollector != nil {
			config := &httptrace.MiddlewareConfig{
				RouteName: func(r *http.Request) string {
					return routeName
				},
				SetContextSpan: func(r *http.Request, id appdash.SpanID) *http.Request {
					ctx := r.Context()
					ctx = traceutil.NewContext(ctx, id)
					ctx = sourcegraph.WithClientMetadata(ctx, (&traceutil.Span{SpanID: id}).Metadata())
					return r.WithContext(ctx)
				},
			}
			m := httptrace.Middleware(traceutil.DefaultCollector, config)
			m(rwIntercept, r, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Appdash-Trace", traceutil.SpanIDFromContext(r.Context()).Trace.String())
				next.ServeHTTP(w, r)
			})
		} else {
			next.ServeHTTP(rwIntercept, r)
		}

		// If the code is zero, the inner Handler never explicitly called
		// WriterHeader. We can assume the response code is 200 in such a case
		code := rwIntercept.Code
		if code == 0 {
			code = 200
		}

		duration := time.Now().Sub(start)
		labels := prometheus.Labels{
			"route":  routeName,
			"method": strings.ToLower(r.Method),
			"code":   strconv.Itoa(code),
			"repo":   repotrackutil.GetTrackedRepo(r.URL.Path),
		}
		requestDuration.With(labels).Observe(duration.Seconds())
		requestHeartbeat.With(labels).Set(float64(time.Now().Unix()))

		uid, sessionID := sessionInfo(r)
		log15.Debug("TRACE HTTP", "method", r.Method, "URL", r.URL.String(), "routename", routeName, "spanID", traceutil.SpanIDFromContext(r.Context()), "code", code, "RemoteAddr", r.RemoteAddr, "UserAgent", r.UserAgent(), "uid", uid, "session", sessionID, "duration", duration)
	})
}

func TraceRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if p, ok := r.Context().Value(routeNameKey).(*string); ok {
			*p = mux.CurrentRoute(r).GetName()
		}
		next.ServeHTTP(rw, r)
	})
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

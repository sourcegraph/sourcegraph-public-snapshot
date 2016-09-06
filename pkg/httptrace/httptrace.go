package httptrace

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
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
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		// -- currently we don't associate XHR calls with the parent page's span --
		// parentSpanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		// if err != nil && err != opentracing.ErrSpanContextNotFound {
		// 	log15.Error("extracting parent span failed", "error", err)
		// }

		// start new span
		span := opentracing.StartSpan("")
		span.SetTag("URL", r.URL.String())
		defer span.Finish()
		rw.Header().Set("X-Trace", traceutil.SpanURL(span))
		ctx = opentracing.ContextWithSpan(ctx, span)
		ctx = traceutil.InjectGRPCMetadata(ctx, span.Context()) // this assumes that the span does not change until any GRPC call, which is a bit bad

		routeName := "unknown"
		ctx = context.WithValue(ctx, routeNameKey, &routeName)

		rwIntercept := &ResponseWriterStatusIntercept{ResponseWriter: rw}
		next.ServeHTTP(rwIntercept, r.WithContext(ctx))

		// route name is only known after the request has been handled
		span.SetOperationName("Serve: " + routeName)
		span.SetTag("Route", routeName)
		span.SetTag("Method", r.Method)
		span.SetTag("URL", r.URL.String())

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

		log15.Debug("TRACE HTTP", "method", r.Method, "URL", r.URL.String(), "routename", routeName, "trace", traceutil.SpanURL(span), "code", code, "RemoteAddr", r.RemoteAddr, "UserAgent", r.UserAgent(), "duration", duration)
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

package handlerutil

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/schema"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/httpwrapper"

	"code.google.com/p/rog-go/parallel"

	"strconv"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
	"src.sourcegraph.com/sourcegraph/util/metricutil"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

var (
	schemaDecoder = schema.NewDecoder()
	once          sync.Once
)

func init() {
	once.Do(func() {
		schemaDecoder.IgnoreUnknownKeys(true)
	})
}

// Handler is the outermost http.Handler wrapper. A request is is handled by the following handlers in order:
//
// 1. Set AuthedUID
// 2. Logging
// 3. Appdash
// 4. Set user object and token
// 5. Run handler, check error resp
func Handler(h HandlerWithErrorReturn) http.Handler {
	mw := []Middleware{logMiddleware}
	if traceutil.DefaultCollector != nil {
		mw = append(mw, httptrace.Middleware(traceutil.DefaultCollector, traceMiddlewareConfig))
	}
	mw = append(mw, httpwrapper.MakeMiddleware(httpwrapperConfig))

	return WithMiddleware(h, mw...)
}

var traceMiddlewareConfig = &httptrace.MiddlewareConfig{
	RouteName: func(r *http.Request) string { return httpctx.RouteName(r) },
	CurrentUser: func(r *http.Request) string {
		a := auth.ActorFromContext(httpctx.FromRequest(r))
		return strconv.Itoa(a.UID)
	},
	SetContextSpan: func(r *http.Request, id appdash.SpanID) {
		traceutil.SetSpanID(r, id)

		ctx := httpctx.FromRequest(r)
		ctx = traceutil.NewContext(ctx, id)
		ctx = sourcegraph.WithClientMetadata(ctx, (&traceutil.Span{SpanID: id}).Metadata())
		httpctx.SetForRequest(r, ctx)
	},
}

var httpwrapperConfig = &httpwrapper.ServerConfig{
	WithActiveSpanFunc: func(r *http.Request, span instrument.ActiveSpan) {
		span.SetName(fmt.Sprintf("http/%s", httpctx.RouteName(r)))
		span.Log(instrument.EventName("cr/span_attributes").Payload(map[string]string{
			"route_path": r.URL.Path,
		}))

		spanID := traceutil.SpanID(r)
		span.Log(instrument.EventName("appdash_span_id").Payload(spanID))
		span.AddTraceJoinId("appdash_trace_id", spanID.Trace)
	},
}

var metricLabels = []string{"route", "method", "code"}
var requestCount = prometheus.NewCounter(stdprometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "requests_total",
	Help:      "Total number of HTTP requests made.",
}, metricLabels)
var requestDuration = prometheus.NewSummary(stdprometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "request_duration_nanoseconds",
	Help:      "The HTTP request latencies in nanoseconds.",
}, metricLabels)
var requestHeartbeat = prometheus.NewGauge(stdprometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "requests_last_timestamp_unixtime",
	Help:      "Last time a request finished for a http endpoint.",
}, metricLabels)

var logMiddleware = func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rwIntercept := &metricutil.ResponseWriterStatusIntercept{ResponseWriter: rw}
	start := time.Now()
	next(rwIntercept, r)

	duration := time.Now().Sub(start)
	routeLabel := metrics.Field{Key: "route", Value: httpctx.RouteName(r)}
	methodLabel := metrics.Field{Key: "method", Value: strings.ToLower(r.Method)}
	successLabel := metrics.Field{Key: "code", Value: strconv.Itoa(rwIntercept.Code)}
	requestCount.With(routeLabel).With(methodLabel).With(successLabel).Add(1)
	requestDuration.With(routeLabel).With(methodLabel).With(successLabel).Observe(duration.Nanoseconds())
	requestHeartbeat.With(routeLabel).With(methodLabel).With(successLabel).Set(float64(time.Now().Unix()))

	log15.Debug("Request", "pkg", "handlerutil", "method", r.Method, "URL", r.URL.String(), "routename", httpctx.RouteName(r), "duration", duration, "code", rwIntercept.Code)
}

// HandlerWithErrorReturn wraps a http.HandlerFunc-like func that also
// returns an error.  If the error is nil, this wrapper is a no-op. If
// the error is non-nil, it attempts to determine the HTTP status code
// equivalent of the returned error (if non-nil) and set that as the
// HTTP status. If a non-nil error is returned
type HandlerWithErrorReturn struct {
	Handler func(http.ResponseWriter, *http.Request) error       // the underlying handler
	Error   func(http.ResponseWriter, *http.Request, int, error) // called to send an error response (e.g., an error page)
}

func (h HandlerWithErrorReturn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		if rv := recover(); rv != nil {
			err = fmt.Errorf("panic: %v", rv)
			log.Println(string(debug.Stack()))
			status := http.StatusInternalServerError
			reportError(r, status, err, true)
			h.Error(w, r, status, err)
		}
	}()

	err = collapseMultipleErrors(h.Handler(w, r))
	if err != nil {
		status := errcode.HTTP(err)
		reportError(r, status, err, false)
		h.Error(w, r, status, err)
	}
}

// collapseMultipleErrors returns the first err if err is a
// parallel.Errors list of length 1. Otherwise it returns err
// unchanged. This lets us return the proper HTTP status code for
// single errors.
func collapseMultipleErrors(err error) error {
	if errs, ok := err.(parallel.Errors); ok && len(errs) == 1 {
		return errs[0]
	}
	return err
}

package trace

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/redact"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/repotrackutil"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type key int

const (
	routeNameKey key = iota
	userKey
	requestErrorCauseKey
	graphQLRequestNameKey
	originKey
	sourceKey
	GraphQLQueryKey
)

// trackOrigin specifies a URL value. When an incoming request has the request header "Origin" set
// and the header value equals the `trackOrigin` value then the `requestDuration` metric (and other metrics downstream)
// gets labeled with this value for the "origin" label  (otherwise the metric is labeled with "unknown").
// The tracked value can be changed with the METRICS_TRACK_ORIGIN environmental variable.
var trackOrigin = "https://gitlab.com"

var (
	metricLabels    = []string{"route", "method", "code", "repo", "origin"}
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_http_request_duration_seconds",
		Help:    "The HTTP request latencies in seconds.",
		Buckets: UserLatencyBuckets,
	}, metricLabels)
)

var requestHeartbeat = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "src_http_requests_last_timestamp_unixtime",
	Help: "Last time a request finished for a http endpoint.",
}, metricLabels)

func Init() {
	if origin := os.Getenv("METRICS_TRACK_ORIGIN"); origin != "" {
		trackOrigin = origin
	}
}

// GraphQLRequestName returns the GraphQL request name for a request context. For example,
// a request to /.api/graphql?Foobar would have the name `Foobar`. If the request had no
// name, or the context is not a GraphQL request, "unknown" is returned.
func GraphQLRequestName(ctx context.Context) string {
	v, ok := ctx.Value(graphQLRequestNameKey).(string)
	if ok {
		return v
	}
	return "unknown"
}

// WithGraphQLRequestName sets the GraphQL request name in the context.
func WithGraphQLRequestName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, graphQLRequestNameKey, name)
}

// RequestOrigin returns the request origin (the value of the request header "Origin") for a request context.
// If the request didn't have this header set "unknown" is returned.
func RequestOrigin(ctx context.Context) string {
	v := ctx.Value(originKey)
	if v == nil {
		return "unknown"
	}
	return v.(string)
}

// WithRequestOrigin sets the request origin in the context.
func WithRequestOrigin(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, originKey, name)
}

// SourceType indicates the type of source that likely created the request.
type SourceType string

const (
	// SourceBrowser indicates the request likely came from a web browser.
	SourceBrowser SourceType = "browser"

	// SourceOther indicates the request likely came from a non-browser HTTP client.
	SourceOther SourceType = "other"
)

// WithRequestSource sets the request source type in the context.
func WithRequestSource(ctx context.Context, source SourceType) context.Context {
	return context.WithValue(ctx, sourceKey, source)
}

// RequestSource returns the request source constant for a request context.
func RequestSource(ctx context.Context) SourceType {
	v := ctx.Value(sourceKey)
	if v == nil {
		return SourceOther
	}
	return v.(SourceType)
}

// slowPaths is a list of endpoints that are slower than the average and for
// which we only want to log a message if the duration is slower than the
// threshold here.
var slowPaths = map[string]time.Duration{
	"/repo-update": 5 * time.Second,
}

var (
	minDuration = env.MustGetDuration("SRC_HTTP_LOG_MIN_DURATION", 2*time.Second, "min duration before slow http requests are logged")
	minCode     = env.MustGetInt("SRC_HTTP_LOG_MIN_CODE", 500, "min http code before http responses are logged")
)

// HTTPMiddleware captures and exports metrics to Prometheus, etc.
//
// 🚨 SECURITY: This handler is served to all clients, even on private servers to clients who have
// not authenticated. It must not reveal any sensitive information.
func HTTPMiddleware(logger log.Logger, next http.Handler, siteConfig conftypes.SiteConfigQuerier) http.Handler {
	logger = logger.Scoped("http", "http tracing middleware")
	return loggingRecoverer(logger, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// extract propagated span
		wireContext, err := ot.GetTracer(ctx).Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			logger.Warn("extracting parent span failed", log.Error(err))
		}

		// start new span
		span, ctx := ot.StartSpanFromContext(ctx, "", ext.RPCServerOption(wireContext))
		ext.HTTPUrl.Set(span, r.URL.String())
		ext.HTTPMethod.Set(span, r.Method)
		span.SetTag("http.referer", r.Header.Get("referer"))
		defer span.Finish()

		// get trace ID
		trace := Context(ctx)
		var traceURL string
		if trace.TraceID != "" {
			traceURL = URL(trace.TraceID, siteConfig)
			rw.Header().Set("X-Trace", traceURL)
			logger = logger.WithTrace(trace)
		}

		ctx = opentracing.ContextWithSpan(ctx, span)

		// route name is only known after the request has been handled
		routeName := "unknown"
		ctx = context.WithValue(ctx, routeNameKey, &routeName)

		var userID int32
		ctx = context.WithValue(ctx, userKey, &userID)

		var requestErrorCause error
		ctx = context.WithValue(ctx, requestErrorCauseKey, &requestErrorCause)

		origin := "unknown"
		if r.Header.Get("Origin") == trackOrigin {
			origin = trackOrigin
		}
		ctx = WithRequestOrigin(ctx, origin)

		// handle request
		m := httpsnoop.CaptureMetrics(next, rw, r.WithContext(ctx))

		// get route name, which is set after request is handled
		if routeName == "graphql" {
			// We use the query to denote the type of a GraphQL request, e.g. /.api/graphql?Repositories
			if r.URL.RawQuery != "" {
				routeName = "graphql: " + r.URL.RawQuery
			} else {
				routeName = "graphql: unknown"
			}
		}
		span.SetOperationName("Serve: " + routeName)
		span.SetTag("Route", routeName)

		ext.HTTPStatusCode.Set(span, uint16(m.Code))

		labels := prometheus.Labels{
			"route":  routeName,
			"method": strings.ToLower(r.Method),
			"code":   strconv.Itoa(m.Code),
			"repo":   repotrackutil.GetTrackedRepo(api.RepoName(r.URL.Path)),
			"origin": origin,
		}
		requestDuration.With(labels).Observe(m.Duration.Seconds())
		requestHeartbeat.With(labels).Set(float64(time.Now().Unix()))

		// if it's not a graphql request, then this includes graphql_error=false in the log entry
		gqlErr := false
		span.Context().ForeachBaggageItem(func(k, v string) bool {
			if k == "graphql.error" {
				gqlErr = true
			}
			return !gqlErr
		})

		if customDuration, ok := slowPaths[r.URL.Path]; ok {
			minDuration = customDuration
		}

		if m.Code >= minCode || m.Duration >= minDuration {
			fields := make([]log.Field, 0, 10)
			fields = append(fields,
				log.String("route_name", routeName),
				log.String("method", r.Method),
				log.String("url", truncate(r.URL.String(), 100)),
				log.Int("code", m.Code),
				log.Duration("duration", m.Duration),
			)

			if v := r.Header.Get("X-Forwarded-For"); v != "" {
				fields = append(fields, log.String("x_forwarded_for", v))
			}

			if userID != 0 {
				fields = append(fields, log.Int("user", int(userID)))
			}

			if gqlErr {
				fields = append(fields, log.Bool("graphql_error", gqlErr))
			}
			var parts []string
			if m.Duration >= minDuration {
				parts = append(parts, "slow http request")
			}
			if m.Code >= minCode {
				parts = append(parts, "unexpected status code")
			}

			msg := strings.Join(parts, ", ")
			switch {
			case m.Code == http.StatusNotFound:
				logger.Info(msg, fields...)
			case m.Code == http.StatusNotAcceptable:
				// Used for intentionally disabled endpoints
				// https://www.rfc-editor.org/rfc/rfc9110.html#name-406-not-acceptable
				logger.Debug(msg, fields...)
			case m.Code == http.StatusUnauthorized:
				logger.Warn(msg, fields...)
			case m.Code >= http.StatusInternalServerError && requestErrorCause != nil:
				// Always wrapping error without a true cause creates loads of events on which we
				// do not have the stack trace and that are barely usable. Once we find a better
				// way to handle such cases, we should bring back the deleted lines from
				// https://github.com/sourcegraph/sourcegraph/pull/29312.
				fields = append(fields, log.Error(requestErrorCause))
				logger.Error(msg, fields...)
			case m.Duration >= minDuration:
				logger.Warn(msg, fields...)
			default:
				logger.Error(msg, fields...)
			}
		}

		// Notify sentry if the status code indicates our system had an error (e.g. 5xx).
		if m.Code >= 500 {
			if requestErrorCause == nil {
				// Always wrapping error without a true cause creates loads of events on which we
				// do not have the stack trace and that are barely usable. Once we find a better
				// way to handle such cases, we should bring back the deleted lines from
				// https://github.com/sourcegraph/sourcegraph/pull/29312.
				return
			}
		}
	}))
}

// Recoverer is a recovery handler to wrap the stdlib net/http Mux.
// Example:
//  mux := http.NewServeMux
//  ...
//	http.Handle("/", sentry.Recoverer(mux))
func loggingRecoverer(logger log.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err := errors.Errorf("handler panic: %v", redact.Safe(r))
				logger.Error("handler panic", log.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler.ServeHTTP(w, r)
	})
}

func truncate(s string, n int) string {
	if len(s) > n {
		return fmt.Sprintf("%s...(%d more)", s[:n], len(s)-n)
	}
	return s
}

func Route(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if p, ok := r.Context().Value(routeNameKey).(*string); ok {
			if routeName := mux.CurrentRoute(r).GetName(); routeName != "" {
				*p = routeName
			}
		}
		next.ServeHTTP(rw, r)
	})
}

func User(ctx context.Context, userID int32) {
	if p, ok := ctx.Value(userKey).(*int32); ok {
		*p = userID
	}
}

// SetRequestErrorCause will set the error for the request to err. This is
// used in the reporting layer to inspect the error for richer reporting to
// Sentry.
func SetRequestErrorCause(ctx context.Context, err error) {
	if p, ok := ctx.Value(requestErrorCauseKey).(*error); ok {
		*p = err
	}
}

// SetRouteName manually sets the name for the route. This should only be used
// for non-mux routed routes (ie middlewares).
func SetRouteName(r *http.Request, routeName string) {
	if p, ok := r.Context().Value(routeNameKey).(*string); ok {
		*p = routeName
	}
}

func WithRouteName(name string, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		SetRouteName(r, name)
		next(rw, r)
	}
}

package trace

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/repotrackutil"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
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
func HTTPMiddleware(next http.Handler, siteConfig conftypes.SiteConfigQuerier) http.Handler {
	return sentry.Recoverer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		wireContext, err := ot.GetTracer(ctx).Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			log15.Error("extracting parent span failed", "error", err)
		}

		// start new span
		span, ctx := ot.StartSpanFromContext(ctx, "", ext.RPCServerOption(wireContext))
		ext.HTTPUrl.Set(span, r.URL.String())
		ext.HTTPMethod.Set(span, r.Method)
		span.SetTag("http.referer", r.Header.Get("referer"))
		defer span.Finish()

		traceID := IDFromSpan(span)
		traceURL := URL(traceID, conf.ExternalURL(), conf.Tracer())

		rw.Header().Set("X-Trace", traceURL)
		ctx = opentracing.ContextWithSpan(ctx, span)

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

		m := httpsnoop.CaptureMetrics(next, rw, r.WithContext(ctx))

		if routeName == "graphql" {
			// We use the query to denote the type of a GraphQL request, e.g. /.api/graphql?Repositories
			if r.URL.RawQuery != "" {
				routeName = "graphql: " + r.URL.RawQuery
			} else {
				routeName = "graphql: unknown"
			}
		}

		// route name is only known after the request has been handled
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

		if m.Duration >= minDuration || m.Code >= minCode {
			kvs := make([]interface{}, 0, 20)
			kvs = append(kvs,
				"method", r.Method,
				"url", truncate(r.URL.String(), 100),
				"code", m.Code,
				"duration", m.Duration,
			)

			if traceID != "" {
				kvs = append(kvs, "traceID", traceID)
			}

			if v := r.Header.Get("X-Forwarded-For"); v != "" {
				kvs = append(kvs, "x_forwarded_for", v)
			}

			if userID != 0 {
				kvs = append(kvs, "user", userID)
			}

			if gqlErr {
				kvs = append(kvs, "graphql_error", gqlErr)
			}
			var parts []string
			if m.Duration >= minDuration {
				parts = append(parts, "slow http request")
			}
			if m.Code >= minCode {
				parts = append(parts, "unexpected status code")
			}
			log15.Warn(strings.Join(parts, ", "), kvs...)
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

			sentry.CaptureError(requestErrorCause, map[string]string{
				"code":            strconv.Itoa(m.Code),
				"method":          r.Method,
				"url":             r.URL.String(),
				"route_name":      routeName,
				"user_agent":      r.UserAgent(),
				"user":            strconv.FormatInt(int64(userID), 10),
				"x_forwarded_for": r.Header.Get("X-Forwarded-For"),
				"written":         strconv.FormatInt(m.Written, 10),
				"duration":        m.Duration.String(),
				"graphql_error":   strconv.FormatBool(gqlErr),
				"trace":           traceURL,
			})
		}
	}))
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

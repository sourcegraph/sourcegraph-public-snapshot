package httptrace

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
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
		uid := "0"
		sessionID := "unknown"
		tok := ""
		start := time.Now()
		rwIntercept := &ResponseWriterStatusIntercept{ResponseWriter: rw}

		sess, err := appauth.ReadSessionCookie(r)
		if err == nil {
			tok = sess.AccessToken
			ctx := r.Context()
			if ctx != nil {
				for _, cookie := range r.Cookies() {
					// each environment uses a different cookie name, but they all start with 'amplitude'
					if strings.Contains(cookie.Name, "amplitude") {
						sessionID = cookie.Value
					}
				}
			}
			ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok, TokenType: "Bearer"}))
			actor, err := accesstoken.ParseAndVerify(idkey.FromContext(ctx), tok)
			if err != nil {
				log15.Debug("Cookie parse:", "errror", err)
			} else {
				uid = strconv.Itoa(actor.UID)
			}
		}

		routeName := "unknown"
		r = r.WithContext(context.WithValue(r.Context(), routeNameKey, &routeName))

		next.ServeHTTP(rwIntercept, r)

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

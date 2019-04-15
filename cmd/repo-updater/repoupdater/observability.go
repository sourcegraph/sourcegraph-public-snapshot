package repoupdater

import (
	"net/http"
	"strconv"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// HandlerMetrics encapsulates the Prometheus metrics of an http.Handler.
type HandlerMetrics struct {
	ServeHTTP *repos.OperationMetrics
}

// NewHandlerMetrics returns HandlerMetrics that need to be registered
// in a Prometheus registry.
func NewHandlerMetrics() HandlerMetrics {
	return HandlerMetrics{
		ServeHTTP: &repos.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "repoupdater",
				Name:      "http_handler_duration_seconds",
				Help:      "Time spent handling an HTTP request",
			}, []string{"path", "code"}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "repoupdater",
				Name:      "http_handler_requests_total",
				Help:      "Total number of HTTP requests",
			}, []string{"path", "code"}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "repoupdater",
				Name:      "http_handler_errors_total",
				Help:      "Total number of HTTP error responses (code >= 400)",
			}, []string{"path", "code"}),
		},
	}
}

// ObservedHandler returns a decorator that wraps an http.Handler
// with logging, Prometheus metrics and tracing.
func ObservedHandler(
	log log15.Logger,
	m HandlerMetrics,
	tr opentracing.Tracer,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return nethttp.Middleware(tr,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rr := &responseRecorder{w, http.StatusOK, 0}
				defer func(begin time.Time) {
					if log != nil {
						log.Info(
							"http.request",
							"method", r.Method,
							"route", r.URL.Path,
							"code", rr.code,
							"duration", time.Since(begin),
						)
					}

					if m.ServeHTTP != nil {
						secs := time.Since(begin).Seconds()
						var err error
						if rr.code >= 400 {
							err = errors.New(http.StatusText(rr.code))
						}
						m.ServeHTTP.Observe(secs, 1, &err, r.URL.Path, strconv.Itoa(rr.code))
					}
				}(time.Now())

				next.ServeHTTP(rr, r)
			}),
		)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	code    int
	written int64
}

// WriteHeader may not be explicitly called, so care must be taken to
// initialize w.code to its default value of http.StatusOK.
func (w *responseRecorder) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseRecorder) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.written += int64(n)
	return n, err
}

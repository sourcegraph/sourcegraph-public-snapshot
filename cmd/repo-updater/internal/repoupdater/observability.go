package repoupdater

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

// HandlerMetrics encapsulates the Prometheus metrics of an http.Handler.
type HandlerMetrics struct {
	ServeHTTP *metrics.REDMetrics
}

// NewHandlerMetrics returns HandlerMetrics that need to be registered
// in a Prometheus registry.
func NewHandlerMetrics() HandlerMetrics {
	return HandlerMetrics{
		ServeHTTP: &metrics.REDMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_http_handler_duration_seconds",
				Help: "Time spent handling an HTTP request",
			}, []string{"path", "code"}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_http_handler_requests_total",
				Help: "Total number of HTTP requests",
			}, []string{"path", "code"}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_http_handler_errors_total",
				Help: "Total number of HTTP error responses (code >= 400)",
			}, []string{"path", "code"}),
		},
	}
}

// MustRegister registers all metrics in HandlerMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (m HandlerMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(m.ServeHTTP.Count)
	r.MustRegister(m.ServeHTTP.Duration)
	r.MustRegister(m.ServeHTTP.Errors)
}

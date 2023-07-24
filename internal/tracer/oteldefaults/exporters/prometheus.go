package exporters

import (
	"github.com/prometheus/client_golang/prometheus"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

// NewPrometheusExporter sets up a metrics Reader for interacting with a
// Prometheus exporter based on prometheus.DefaultRegisterer
func NewPrometheusExporter() (metricsdk.Reader, error) {
	return otelprometheus.New(
		otelprometheus.WithRegisterer(prometheus.DefaultRegisterer))
}

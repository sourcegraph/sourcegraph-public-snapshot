package opentelemetry

import (
	"context"
	"time"

	gcpmetricexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func configureMetrics(_ context.Context, logger log.Logger, config Config, res *resource.Resource) (func(), error) {
	var reader sdkmetric.Reader
	if config.GCPProjectID != "" {
		logger.Info("initializing GCP metric exporter", log.String("projectID", config.GCPProjectID))
		// We use a push-based exporter with PeriodicReader from the metrics SDK
		// to publish metrics to GCP.
		exporter, err := gcpmetricexporter.New(
			gcpmetricexporter.WithProjectID(config.GCPProjectID))
		if err != nil {
			return nil, errors.Wrap(err, "gcpmetricexporter.New")
		}
		reader = sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(30*time.Second))
	} else {
		logger.Info("initializing Prometheus exporter")
		// We use a pull-based exporter instead for Prometheus, where a Prometheus
		// instance can optionally be configured to pull from this service.
		var err error
		reader, err = otelprometheus.New(
			otelprometheus.WithRegisterer(prometheus.DefaultRegisterer))
		if err != nil {
			return nil, errors.Wrap(err, "exporters.NewPrometheusExporter")
		}
	}

	// Initialize instrumentation
	collectedMetrics, err := meter.Int64Counter("otel.metric.collected.count",
		metric.WithDescription("Total number of metrics collected, and whether the collection succeeded or failed."))
	if err != nil {
		return nil, errors.Wrap(err, "initialize collectedMetrics metric")
	}

	// Create and set global tracer, wrapping the metric reader in instrumentation
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(&instrumentedMetricReader{
			Reader: reader,
			// NOTE: This does NOT work with the Prometheus exporter, because it
			// wraps an internal reader, and the top-level Collect does not get
			// called unless an OpenTelemetry metric exporter is used. If we want
			// to add an equivalent for Prometheus exporter, we probably need
			// to wrap the prometheus Gatherer implementation instead, but since
			// it's mostly intended for local dev, this is fine to not do for now.
			collectedMetrics: collectedMetrics,
		}),
		sdkmetric.WithResource(res))
	otel.SetMeterProvider(provider)

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		logger.Debug("shutting down metrics")
		if err := provider.ForceFlush(shutdownCtx); err != nil {
			logger.Warn("error occurred force-flushing metrics", log.Error(err))
		}
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.Warn("error occurred shutting down metrics", log.Error(err))
		}
		logger.Info("metrics shut down", log.Duration("elapsed", time.Since(start)))
	}, nil
}

// instrumentedMetricReader wraps sdkmetric.Reader in instrumentation because the
// SDK does not provide metrics out of the box. This effectively measures success
// and volume of metrics export.
type instrumentedMetricReader struct {
	sdkmetric.Reader

	collectedMetrics metric.Int64Counter
}

var _ sdkmetric.Reader = &instrumentedMetricReader{}

func (r *instrumentedMetricReader) Collect(ctx context.Context, data *metricdata.ResourceMetrics) error {
	err := r.Reader.Collect(ctx, data) // call underlying first

	if data != nil { // guard out of abundance of caution
		var count int
		for _, s := range data.ScopeMetrics {
			count += len(s.Metrics)
		}
		r.collectedMetrics.Add(ctx, int64(count), metric.WithAttributeSet(attribute.NewSet(
			attribute.Bool("succeeded", err == nil),
		)))
	}

	return err
}

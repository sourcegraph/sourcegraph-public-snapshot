package opentelemetry

import (
	"context"
	"time"

	gcpmetricexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	apioption "google.golang.org/api/option"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func configureMetrics(_ context.Context, logger log.Logger, config Config, res *resource.Resource) (func(), error) {
	var reader sdkmetric.Reader
	if config.GCPProjectID != "" {
		logger.Info("initializing GCP metric exporter", log.String("projectID", config.GCPProjectID))
		// We use a push-based exporter with PeriodicReader from the metrics SDK
		// to publish metrics to GCP.
		exporter, err := gcpmetricexporter.New(
			gcpmetricexporter.WithProjectID(config.GCPProjectID),
			gcpmetricexporter.WithMonitoringClientOptions(
				apioption.WithTelemetryDisabled(),
			))
		if err != nil {
			return nil, errors.Wrap(err, "gcpmetricexporter.New")
		}

		// Initialize instrumentation
		collectedMetrics, err := meter.Int64Counter("otel.metric.exported.count",
			metric.WithDescription("Total number of metrics collected, and whether the collection succeeded or failed."))
		if err != nil {
			return nil, errors.Wrap(err, "initialize collectedMetrics metric")
		}

		reader = sdkmetric.NewPeriodicReader(
			// Wrap exporter in instrumentation
			&instrumentedMetricExporter{
				Exporter:         exporter,
				collectedMetrics: collectedMetrics,
			},
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
		// Note: no easy way to instrument here, as we just get a Reader, which
		// isn't the right place to instrument exports. Since this is mostly
		// local dev, just ignore for now
	}

	// Create and set global tracer
	provider := sdkmetric.NewMeterProvider(
		// Note: adding reader-level instrumentation here doesn't seem to be
		// the right way to collect metrics, it must be added to sdkmetric.Exporter
		// implementations instead.
		sdkmetric.WithReader(reader),
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

// instrumentedMetricExporter wraps sdkmetric.Exporter in instrumentation because
// the SDK does not provide metrics out of the box.
type instrumentedMetricExporter struct {
	sdkmetric.Exporter

	collectedMetrics metric.Int64Counter
}

var _ sdkmetric.Exporter = &instrumentedMetricExporter{}

func (r *instrumentedMetricExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	err := r.Exporter.Export(ctx, data) // call underlying first

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

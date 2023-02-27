package otlpadapter

import (
	"context"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/internal/otlpenv"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Register sets up adapter services and registers proxies on the router. enabled can be
// provided to atomically toggle whether the signal endpoints are enabled serverside -
// this is important because clients might not receive updated configuration for quite a
// while, so we need to stop accepting incoming requests.
func Register(ctx context.Context, logger log.Logger, protocol otlpenv.Protocol, endpoint string, r *mux.Router, enabled *atomic.Bool) {
	// Build an OTLP exporter that exports directly to the desired protocol and endpoint
	exporterFactory, signalExporterConfig, err := newExporter(protocol, endpoint)
	if err != nil {
		logger.Fatal("newExporter", log.Error(err))
	}

	// Receive OTLP http/json signals
	receiverURL, err := url.Parse("http://127.0.0.1:4319")
	if err != nil {
		logger.Fatal("unreachable", log.Error(err))
	}
	receiverFactory, signalReceiverConfig := newReceiver(receiverURL)

	// Set up shared configuration for creating signal exporters and receivers. Telemetry
	// settigns are required on all factories, and all fields of this struct are required.
	telemetrySettings := component.TelemetrySettings{
		Logger: zap.NewNop(),

		TracerProvider: otel.GetTracerProvider(),

		MeterProvider: mustPrometheusMetricsProvider(logger),
		MetricsLevel:  configtelemetry.LevelBasic,
	}
	componentName := "otlpadapter"

	// otelSignals declares the signals we support redirection for.
	var otelSignals = []adaptedSignal{
		{
			PathPrefix: "/v1/traces",
			CreateAdapter: func() (*signalAdapter, error) {
				exporter, err := exporterFactory.CreateTracesExporter(ctx, exporter.CreateSettings{
					ID:                component.NewIDWithName(component.DataTypeTraces, componentName),
					TelemetrySettings: telemetrySettings,
				}, signalExporterConfig)
				if err != nil {
					return nil, errors.Wrap(err, "CreateTracesExporter")
				}
				receiver, err := receiverFactory.CreateTracesReceiver(ctx, receiver.CreateSettings{
					ID:                component.NewIDWithName(component.DataTypeTraces, componentName),
					TelemetrySettings: telemetrySettings,
				}, signalReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrap(err, "CreateTracesReceiver")
				}
				return &signalAdapter{Exporter: exporter, Receiver: receiver}, nil
			},
			Enabled: enabled,
		},
		{
			PathPrefix: "/v1/metrics",
			CreateAdapter: func() (*signalAdapter, error) {
				exporter, err := exporterFactory.CreateMetricsExporter(ctx, exporter.CreateSettings{
					ID:                component.NewIDWithName(component.DataTypeMetrics, componentName),
					TelemetrySettings: telemetrySettings,
				}, signalExporterConfig)
				if err != nil {
					return nil, errors.Wrap(err, "CreateMetricsExporter")
				}
				receiver, err := receiverFactory.CreateMetricsReceiver(ctx, receiver.CreateSettings{
					ID:                component.NewIDWithName(component.DataTypeMetrics, componentName),
					TelemetrySettings: telemetrySettings,
				}, signalReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrap(err, "CreateMetricsReceiver")
				}
				return &signalAdapter{Exporter: exporter, Receiver: receiver}, nil
			},
			Enabled: enabled,
		},
		{
			PathPrefix: "/v1/logs",
			CreateAdapter: func() (*signalAdapter, error) {
				exporter, err := exporterFactory.CreateLogsExporter(ctx, exporter.CreateSettings{
					ID:                component.NewIDWithName(component.DataTypeLogs, componentName),
					TelemetrySettings: telemetrySettings,
				}, signalExporterConfig)
				if err != nil {
					return nil, errors.Wrap(err, "CreateLogsExporter")
				}
				receiver, err := receiverFactory.CreateLogsReceiver(ctx, receiver.CreateSettings{
					ID:                component.NewIDWithName(component.DataTypeLogs, componentName),
					TelemetrySettings: telemetrySettings,
				}, signalReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrap(err, "CreateLogsReceiver")
				}
				return &signalAdapter{Exporter: exporter, Receiver: receiver}, nil
			},
			Enabled: enabled,
		},
	}

	// Finally, spin up redirectors for each signal and set up the appropriate endpoints.
	for _, otelSignal := range otelSignals {
		otelSignal := otelSignal // copy
		otelSignal.Register(ctx, logger, r, receiverURL)
	}
}

// mustPrometheusMetricsProvider creates a metric provider that uses the default
// Prometheus metrics registerer to report metrics. If it fails to create one,
// an error is logged and a no-op metrics provider is used.
func mustPrometheusMetricsProvider(l log.Logger) metric.MeterProvider {
	prom, err := otelprometheus.New(otelprometheus.WithRegisterer(prometheus.DefaultRegisterer))
	if err != nil {
		l.Error("failed to register prometheus metrics for otlpadapter", log.Error(err))
		return metric.NewNoopMeterProvider()
	}
	return sdkmetric.NewMeterProvider(sdkmetric.WithReader(prom))
}

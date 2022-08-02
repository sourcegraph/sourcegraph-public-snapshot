package otlpadapter

import (
	"context"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/internal/otlpenv"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Register sets up adapter services and registers proxies on the router.
func Register(ctx context.Context, logger log.Logger, protocol otlpenv.Protocol, endpoint string, r *mux.Router) {
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

	// Set up shared configuration for creating signal exporters and receivers - telemetry
	// fields are required.
	var (
		signalExporterCreateSettings = component.ExporterCreateSettings{
			TelemetrySettings: component.TelemetrySettings{
				Logger:         zap.NewNop(),
				TracerProvider: otel.GetTracerProvider(),
			},
		}
		signalReceiverCreateSettings = component.ReceiverCreateSettings{
			TelemetrySettings: component.TelemetrySettings{
				Logger:         zap.NewNop(),
				TracerProvider: otel.GetTracerProvider(),
			},
		}
	)

	// otelSignals declares the signals we support redirection for.
	var otelSignals = []adaptedSignal{
		{
			PathPrefix: "/v1/traces",
			CreateAdapter: func() (*signalAdapter, error) {
				exporter, err := exporterFactory.CreateTracesExporter(ctx, signalExporterCreateSettings, signalExporterConfig)
				if err != nil {
					return nil, errors.Wrap(err, "CreateTracesExporter")
				}
				receiver, err := receiverFactory.CreateTracesReceiver(ctx, signalReceiverCreateSettings, signalReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrap(err, "CreateTracesReceiver")
				}
				return &signalAdapter{Exporter: exporter, Receiver: receiver}, nil
			},
		},
		{
			PathPrefix: "/v1/metrics",
			CreateAdapter: func() (*signalAdapter, error) {
				exporter, err := exporterFactory.CreateMetricsExporter(ctx, signalExporterCreateSettings, signalExporterConfig)
				if err != nil {
					return nil, errors.Wrap(err, "CreateMetricsExporter")
				}
				receiver, err := receiverFactory.CreateMetricsReceiver(ctx, signalReceiverCreateSettings, signalReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrap(err, "CreateMetricsReceiver")
				}
				return &signalAdapter{Exporter: exporter, Receiver: receiver}, nil
			},
		},
		{
			PathPrefix: "/v1/logs",
			CreateAdapter: func() (*signalAdapter, error) {
				exporter, err := exporterFactory.CreateLogsExporter(ctx, signalExporterCreateSettings, signalExporterConfig)
				if err != nil {
					return nil, errors.Wrap(err, "CreateLogsExporter")
				}
				receiver, err := receiverFactory.CreateLogsReceiver(ctx, signalReceiverCreateSettings, signalReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrap(err, "CreateLogsReceiver")
				}
				return &signalAdapter{Exporter: exporter, Receiver: receiver}, nil
			},
		},
	}

	// Finally, spin up redirectors for each signal and set up the appropriate endpoints.
	for _, otelSignal := range otelSignals {
		otelSignal := otelSignal // copy
		otelSignal.Register(ctx, logger, r, receiverURL)
	}
}

package otlpadapter

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/std"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/internal/otlpenv"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Register sets up adapter services and registers proxies on the router.
func Register(ctx context.Context, logger log.Logger, protocol, endpoint string, r *mux.Router) {
	// Build an OTLP exporter that exports directly to the desired protocol and endpoint
	var (
		exporterFactory      component.ExporterFactory
		signalExporterConfig config.Exporter
	)
	switch protocol {
	case "grpc":
		exporterFactory = otlpexporter.NewFactory()
		config := exporterFactory.CreateDefaultConfig().(*otlpexporter.Config)
		config.GRPCClientSettings.Endpoint = endpoint
		config.GRPCClientSettings.TLSSetting = configtls.TLSClientSetting{
			Insecure: otlpenv.IsInsecure(endpoint),
		}
		signalExporterConfig = config
	case "http/json":
		exporterFactory = otlphttpexporter.NewFactory()
		config := exporterFactory.CreateDefaultConfig().(*otlphttpexporter.Config)
		config.HTTPClientSettings.Endpoint = endpoint
		signalExporterConfig = config
	default:
		logger.Fatal("unexpected protocol", log.String("protocol", protocol))
	}

	// Receive OTLP http/json signals
	receiverEndpoint, err := url.Parse("http://127.0.0.1:4319")
	if err != nil {
		logger.Fatal("unreachable", log.Error(err))
	}
	receiverFactory := otlpreceiver.NewFactory()
	signalReceiverConfig := receiverFactory.CreateDefaultConfig().(*otlpreceiver.Config)
	signalReceiverConfig.GRPC = nil // disable gRPC receiver, we don't need it
	signalReceiverConfig.HTTP = &confighttp.HTTPServerSettings{
		Endpoint: receiverEndpoint.Host,
	}

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
	}

	// Finally, spin up redirectors for each signal and set up the appropriate endpoints.
	for _, otelSignal := range otelSignals {
		otelSignal := otelSignal // copy lol

		adapterLogger := logger.Scoped(path.Base(otelSignal.PathPrefix), "OpenTelemetry signal-specific tunnel")

		// Set up an http/json -> ${configured_protocol} adapter
		adapter, err := otelSignal.CreateAdapter()
		if err != nil {
			adapterLogger.Fatal("CreateAdapter", log.Error(err))
		}
		if err := adapter.Start(ctx, &otelHost{logger: logger}); err != nil {
			adapterLogger.Fatal("adapter.Start", log.Error(err))
		}

		// The redirector starts up a receiver service running at receiverEndpoint,
		// so now we have to reverse-proxy incoming requests to it so that things get
		// exported correctly.
		r.PathPrefix("/otlp" + otelSignal.PathPrefix).Handler(&httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = receiverEndpoint.Scheme
				req.URL.Host = receiverEndpoint.Host
				req.URL.Path = otelSignal.PathPrefix
			},
			ErrorLog: std.NewLogger(adapterLogger, log.LevelWarn),
		})

		adapterLogger.Info("signal adapter registered")
	}
}

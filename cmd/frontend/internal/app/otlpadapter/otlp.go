package otlpadapter

import (
	"net/url"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/sourcegraph/sourcegraph/internal/otlpenv"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newExporter(
	protocol otlpenv.Protocol,
	endpoint string,
) (
	exporterFactory exporter.Factory,
	signalExporterConfig component.Config,
	err error,
) {
	switch protocol {
	case otlpenv.ProtocolGRPC:
		exporterFactory = otlpexporter.NewFactory()
		tempConfig := exporterFactory.CreateDefaultConfig().(*otlpexporter.Config)
		tempConfig.GRPCClientSettings.Endpoint = endpoint
		tempConfig.GRPCClientSettings.TLSSetting = configtls.TLSClientSetting{
			Insecure: otlpenv.IsInsecure(endpoint),
		}
		signalExporterConfig = tempConfig

	case otlpenv.ProtocolHTTPJSON:
		exporterFactory = otlphttpexporter.NewFactory()
		tempConfig := exporterFactory.CreateDefaultConfig().(*otlphttpexporter.Config)
		tempConfig.HTTPClientSettings.Endpoint = endpoint
		signalExporterConfig = tempConfig

	default:
		err = errors.Newf("unexpected protocol %q", protocol)
	}

	return
}

func newReceiver(receiverURL *url.URL) (receiver.Factory, component.Config) {
	receiverFactory := otlpreceiver.NewFactory()
	signalReceiverConfig := receiverFactory.CreateDefaultConfig().(*otlpreceiver.Config)
	signalReceiverConfig.GRPC = nil // disable gRPC receiver, we don't need it
	signalReceiverConfig.HTTP = &confighttp.HTTPServerSettings{
		Endpoint: receiverURL.Host,
	}

	return receiverFactory, signalReceiverConfig
}

pbckbge otlpbdbpter

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

	"github.com/sourcegrbph/sourcegrbph/internbl/otlpenv"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func newExporter(
	protocol otlpenv.Protocol,
	endpoint string,
) (
	exporterFbctory exporter.Fbctory,
	signblExporterConfig component.Config,
	err error,
) {
	switch protocol {
	cbse otlpenv.ProtocolGRPC:
		exporterFbctory = otlpexporter.NewFbctory()
		tempConfig := exporterFbctory.CrebteDefbultConfig().(*otlpexporter.Config)
		tempConfig.GRPCClientSettings.Endpoint = endpoint
		tempConfig.GRPCClientSettings.TLSSetting = configtls.TLSClientSetting{
			Insecure: otlpenv.IsInsecure(endpoint),
		}
		signblExporterConfig = tempConfig

	cbse otlpenv.ProtocolHTTPJSON:
		exporterFbctory = otlphttpexporter.NewFbctory()
		tempConfig := exporterFbctory.CrebteDefbultConfig().(*otlphttpexporter.Config)
		tempConfig.HTTPClientSettings.Endpoint = endpoint
		signblExporterConfig = tempConfig

	defbult:
		err = errors.Newf("unexpected protocol %q", protocol)
	}

	return
}

func newReceiver(receiverURL *url.URL) (receiver.Fbctory, component.Config) {
	receiverFbctory := otlpreceiver.NewFbctory()
	signblReceiverConfig := receiverFbctory.CrebteDefbultConfig().(*otlpreceiver.Config)
	signblReceiverConfig.GRPC = nil // disbble gRPC receiver, we don't need it
	signblReceiverConfig.HTTP = &confighttp.HTTPServerSettings{
		Endpoint: receiverURL.Host,
	}

	return receiverFbctory, signblReceiverConfig
}

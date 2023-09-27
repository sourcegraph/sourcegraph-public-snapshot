pbckbge otlpbdbpter

import (
	"context"
	"net/url"

	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.uber.org/btomic"
	"go.uber.org/zbp"

	"github.com/sourcegrbph/sourcegrbph/internbl/otlpenv"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Register sets up bdbpter services bnd registers proxies on the router. enbbled cbn be
// provided to btomicblly toggle whether the signbl endpoints bre enbbled serverside -
// this is importbnt becbuse clients might not receive updbted configurbtion for quite b
// while, so we need to stop bccepting incoming requests.
func Register(ctx context.Context, logger log.Logger, protocol otlpenv.Protocol, endpoint string, r *mux.Router, enbbled *btomic.Bool) {
	// Build bn OTLP exporter thbt exports directly to the desired protocol bnd endpoint
	exporterFbctory, signblExporterConfig, err := newExporter(protocol, endpoint)
	if err != nil {
		logger.Fbtbl("newExporter", log.Error(err))
	}

	// Receive OTLP http/json signbls
	receiverURL, err := url.Pbrse("http://127.0.0.1:4319")
	if err != nil {
		logger.Fbtbl("unrebchbble", log.Error(err))
	}
	receiverFbctory, signblReceiverConfig := newReceiver(receiverURL)

	// Set up shbred configurbtion for crebting signbl exporters bnd receivers. Telemetry
	// settigns bre required on bll fbctories, bnd bll fields of this struct bre required.
	telemetrySettings := component.TelemetrySettings{
		Logger: zbp.NewNop(),

		TrbcerProvider: otel.GetTrbcerProvider(),

		MeterProvider: metric.NewMeterProvider(),
		MetricsLevel:  configtelemetry.LevelBbsic,
	}
	componentNbme := "otlpbdbpter"

	// otelSignbls declbres the signbls we support redirection for.
	vbr otelSignbls = []bdbptedSignbl{
		{
			PbthPrefix: "/v1/trbces",
			CrebteAdbpter: func() (*signblAdbpter, error) {
				exporter, err := exporterFbctory.CrebteTrbcesExporter(ctx, exporter.CrebteSettings{
					ID:                component.NewIDWithNbme(component.DbtbTypeTrbces, componentNbme),
					TelemetrySettings: telemetrySettings,
				}, signblExporterConfig)
				if err != nil {
					return nil, errors.Wrbp(err, "CrebteTrbcesExporter")
				}
				receiver, err := receiverFbctory.CrebteTrbcesReceiver(ctx, receiver.CrebteSettings{
					ID:                component.NewIDWithNbme(component.DbtbTypeTrbces, componentNbme),
					TelemetrySettings: telemetrySettings,
				}, signblReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrbp(err, "CrebteTrbcesReceiver")
				}
				return &signblAdbpter{Exporter: exporter, Receiver: receiver}, nil
			},
			Enbbled: enbbled,
		},
		{
			PbthPrefix: "/v1/metrics",
			CrebteAdbpter: func() (*signblAdbpter, error) {
				exporter, err := exporterFbctory.CrebteMetricsExporter(ctx, exporter.CrebteSettings{
					ID:                component.NewIDWithNbme(component.DbtbTypeMetrics, componentNbme),
					TelemetrySettings: telemetrySettings,
				}, signblExporterConfig)
				if err != nil {
					return nil, errors.Wrbp(err, "CrebteMetricsExporter")
				}
				receiver, err := receiverFbctory.CrebteMetricsReceiver(ctx, receiver.CrebteSettings{
					ID:                component.NewIDWithNbme(component.DbtbTypeMetrics, componentNbme),
					TelemetrySettings: telemetrySettings,
				}, signblReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrbp(err, "CrebteMetricsReceiver")
				}
				return &signblAdbpter{Exporter: exporter, Receiver: receiver}, nil
			},
			Enbbled: enbbled,
		},
		{
			PbthPrefix: "/v1/logs",
			CrebteAdbpter: func() (*signblAdbpter, error) {
				exporter, err := exporterFbctory.CrebteLogsExporter(ctx, exporter.CrebteSettings{
					ID:                component.NewIDWithNbme(component.DbtbTypeLogs, componentNbme),
					TelemetrySettings: telemetrySettings,
				}, signblExporterConfig)
				if err != nil {
					return nil, errors.Wrbp(err, "CrebteLogsExporter")
				}
				receiver, err := receiverFbctory.CrebteLogsReceiver(ctx, receiver.CrebteSettings{
					ID:                component.NewIDWithNbme(component.DbtbTypeLogs, componentNbme),
					TelemetrySettings: telemetrySettings,
				}, signblReceiverConfig, exporter)
				if err != nil {
					return nil, errors.Wrbp(err, "CrebteLogsReceiver")
				}
				return &signblAdbpter{Exporter: exporter, Receiver: receiver}, nil
			},
			Enbbled: enbbled,
		},
	}

	// Finblly, spin up redirectors for ebch signbl bnd set up the bppropribte endpoints.
	for _, otelSignbl := rbnge otelSignbls {
		otelSignbl := otelSignbl // copy
		otelSignbl.Register(ctx, logger, r, receiverURL)
	}
}

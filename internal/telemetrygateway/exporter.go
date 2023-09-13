package telemetrygateway

import (
	"context"
	"net/url"

	"google.golang.org/grpc"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Exporter interface {
	ExportEvents(context.Context, []*telemetrygatewayv1.Event) error
}

var address = env.Get("SRC_TELEMETRY_GATEWAY_ADDR", "dns:telemetry-gateway.sourcegraph.com",
	"Target Telemetry Gateway address: https://github.com/grpc/grpc/blob/master/doc/naming.md")

func NewExporter(ctx context.Context, logger log.Logger, c conftypes.SiteConfigQuerier) (Exporter, error) {
	if address == "" {
		return noopExporter{}, nil
	}

	u, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrap(err, "invalid SRC_TELEMETRY_GATEWAY_ADDR")
	}

	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	insecureTarget := u.Scheme != "dns"
	if insecureTarget && !env.InsecureDev {
		return nil, errors.Wrap(err, "insecure SRC_TELEMETRY_GATEWAY_ADDR used outside of dev mode")
	}

	var opts []grpc.DialOption
	if insecureTarget {
		opts = defaults.DialOptions(logger, grpc.WithPerRPCCredentials(&perRPCCredentials{
			conf:     c,
			insecure: true,
		}))
	} else {
		opts = defaults.ExternalDialOptions(logger, grpc.WithPerRPCCredentials(&perRPCCredentials{conf: c}))
	}
	conn, err := grpc.DialContext(ctx, u.String(), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "dialing telemetry gateway")
	}

	return &exporter{client: telemetrygatewayv1.NewTelemeteryGatewayServiceClient(conn)}, nil
}

type noopExporter struct{}

func (e noopExporter) ExportEvents(context.Context, []*telemetrygatewayv1.Event) error { return nil }

type exporter struct {
	client telemetrygatewayv1.TelemeteryGatewayServiceClient
}

func (e *exporter) ExportEvents(ctx context.Context, events []*telemetrygatewayv1.Event) error {
	stream, err := e.client.RecordEvents(ctx)
	if err != nil {
		return errors.Wrap(err, "start export")
	}

	// Send metadata
	if err := stream.Send(&telemetrygatewayv1.RecordEventsRequest{
		Payload: &telemetrygatewayv1.RecordEventsRequest_Metadata{
			Metadata: &telemetrygatewayv1.RecordEventsRequestMetadata{}, // TODO
		},
	}); err != nil {
		return errors.Wrap(err, "send metadata")
	}

	// Start streaming our set of events
	for _, e := range events {
		if err := stream.Send(&telemetrygatewayv1.RecordEventsRequest{
			Payload: &telemetrygatewayv1.RecordEventsRequest_Event{
				Event: e,
			},
		}); err != nil {
			// TODO: Retry elegantly instead of hard-failing
			return errors.Wrap(err, "send event")
		}
	}

	return nil
}

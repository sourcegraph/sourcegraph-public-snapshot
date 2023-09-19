package telemetrygateway

import (
	"context"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/grpc/chunk"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Exporter interface {
	ExportEvents(context.Context, []*telemetrygatewayv1.Event) ([]string, error)
	Close() error
}

func NewExporter(ctx context.Context, logger log.Logger, c conftypes.SiteConfigQuerier, exportAddress string) (Exporter, error) {
	u, err := url.Parse(exportAddress)
	if err != nil {
		return nil, errors.Wrap(err, "invalid export address")
	}

	insecureTarget := u.Scheme != "https"
	if insecureTarget && !env.InsecureDev {
		return nil, errors.Wrap(err, "insecure export address used outside of dev mode")
	}

	var opts []grpc.DialOption
	if insecureTarget {
		opts = defaults.DialOptions(logger)
	} else {
		opts = defaults.ExternalDialOptions(logger)
	}
	conn, err := grpc.DialContext(ctx, u.Host, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "dialing telemetry gateway")
	}

	return &exporter{
		client: telemetrygatewayv1.NewTelemeteryGatewayServiceClient(conn),
		conf:   c,
		conn:   conn,
	}, nil
}

type exporter struct {
	client telemetrygatewayv1.TelemeteryGatewayServiceClient
	conf   conftypes.SiteConfigQuerier
	conn   *grpc.ClientConn
}

func (e *exporter) ExportEvents(ctx context.Context, events []*telemetrygatewayv1.Event) ([]string, error) {
	// Start the stream
	stream, err := e.client.RecordEvents(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start export")
	}

	// Send initial metadata
	if err := stream.Send(&telemetrygatewayv1.RecordEventsRequest{
		Payload: &telemetrygatewayv1.RecordEventsRequest_Metadata{
			Metadata: &telemetrygatewayv1.RecordEventsRequestMetadata{
				LicenseKey: e.conf.SiteConfig().LicenseKey,
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "send initial metadata")
	}

	// Start streaming our set of events, chunking them based on message size
	// as determined internally by chunk.Chunker. Keep track of the events we
	// submit.
	succeededEvents := make([]string, 0, len(events))
	chunker := chunk.New(func(chunkedEvents []*telemetrygatewayv1.Event) error {
		err := stream.Send(&telemetrygatewayv1.RecordEventsRequest{
			Payload: &telemetrygatewayv1.RecordEventsRequest_Events{
				Events: &telemetrygatewayv1.RecordEventsRequest_EventsPayload{
					Events: chunkedEvents,
				},
			},
		})
		failedEvents := make(map[string]struct{})
		if err != nil {
			if details := extractErrorDetails(err); details != nil {
				for _, ev := range details.FailedEvents {
					failedEvents[ev.EventId] = struct{}{}
				}
			}
		}
		for _, ev := range chunkedEvents {
			if _, failed := failedEvents[ev.GetId()]; !failed {
				succeededEvents = append(succeededEvents, ev.Id)
			}
		}
		return err
	})

	if err := chunker.Send(events...); err != nil {
		return succeededEvents, errors.Wrap(err, "chunk and send events")
	}
	if err := chunker.Flush(); err != nil {
		return succeededEvents, errors.Wrap(err, "flush events")
	}

	return succeededEvents, nil
}

func (e *exporter) Close() error { return e.conn.Close() }

func extractErrorDetails(err error) *telemetrygatewayv1.RecordEventsErrorDetails {
	st, ok := status.FromError(err)
	if ok {
		for _, detail := range st.Details() {
			switch d := detail.(type) {
			case *telemetrygatewayv1.RecordEventsErrorDetails:
				return d
			}
		}
	}
	return nil
}

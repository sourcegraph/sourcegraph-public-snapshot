package telemetrygateway

import (
	"context"
	"net/url"

	"google.golang.org/grpc"

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
	defer func() { _ = stream.CloseSend() }()

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
	allSucceededEvents := make([]string, 0, len(events))
	chunker := chunk.New(func(chunkedEvents []*telemetrygatewayv1.Event) error {
		err := stream.Send(&telemetrygatewayv1.RecordEventsRequest{
			Payload: &telemetrygatewayv1.RecordEventsRequest_Events{
				Events: &telemetrygatewayv1.RecordEventsRequest_EventsPayload{
					Events: chunkedEvents,
				},
			},
		})

		if err != nil {
			// If the batch failed, check if we got details about the failure, which
			// we can use to check if our failure is partial or total.
			if chunkSucceeded := getSucceededEventsInError(chunkedEvents, err); len(chunkSucceeded) > 0 {
				allSucceededEvents = append(allSucceededEvents, chunkSucceeded...)
			}
		} else {
			// Otherwise, all events succeeded!
			for _, event := range chunkedEvents {
				allSucceededEvents = append(allSucceededEvents, event.GetId())
			}
		}

		return err
	})

	if err := chunker.Send(events...); err != nil {
		return allSucceededEvents, errors.Wrap(err, "chunk and send events")
	}
	if err := chunker.Flush(); err != nil {
		return allSucceededEvents, errors.Wrap(err, "flush events")
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return allSucceededEvents, errors.Wrap(err, "close request")
	}

	return allSucceededEvents, nil
}

func (e *exporter) Close() error { return e.conn.Close() }

package telemetrygateway

import (
	"context"
	"io"
	"net/url"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/grpc/chunk"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Exporter interface {
	ExportEvents(context.Context, []*telemetrygatewayv1.Event) ([]string, error)
	Close() error
}

func NewExporter(
	ctx context.Context,
	logger log.Logger,
	c conftypes.SiteConfigQuerier,
	g database.GlobalStateStore,
	exportURL *url.URL,
) (Exporter, error) {
	insecureTarget := exportURL.Scheme != "https"
	if insecureTarget && !env.InsecureDev {
		return nil, errors.New("insecure export address used outside of dev mode")
	}

	// TODO(@bobheadxi): Maybe don't use defaults.DialOptions etc, which are
	// geared towards in-Sourcegraph services.
	var opts []grpc.DialOption
	if insecureTarget {
		opts = defaults.DialOptions(logger)
	} else {
		opts = defaults.ExternalDialOptions(logger)
	}
	conn, err := grpc.DialContext(ctx, exportURL.Host, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "dialing telemetry gateway")
	}

	return &exporter{
		client: telemetrygatewayv1.NewTelemeteryGatewayServiceClient(conn),
		conn:   conn,

		globalState: g,
		conf:        c,
	}, nil
}

type exporter struct {
	client telemetrygatewayv1.TelemeteryGatewayServiceClient
	conn   *grpc.ClientConn

	conf        conftypes.SiteConfigQuerier
	globalState database.GlobalStateStore
}

func (e *exporter) ExportEvents(ctx context.Context, events []*telemetrygatewayv1.Event) ([]string, error) {
	tr, ctx := trace.New(ctx, "ExportEvents", attribute.Int("events", len(events)))
	defer tr.End()

	identifier, err := newIdentifier(ctx, e.conf, e.globalState)
	if err != nil {
		tr.SetError(err)
		return nil, err
	}

	var requestID string
	if tr.IsRecording() {
		requestID = tr.SpanContext().TraceID().String()
	} else {
		requestID = uuid.NewString()
	}

	succeeded, err := e.doExportEvents(ctx, requestID, identifier, events)
	if err != nil {
		tr.SetError(err)
		// Surface request ID to help us correlate log entries more easily on
		// our end, because Telemetry Gateway doesn't return granular failure
		// details.
		return succeeded, errors.Wrapf(err, "request %q", requestID)
	}
	return succeeded, nil
}

// doExportEvents makes it easier for us to wrap all errors in our request ID
// for ease of investigating failures.
func (e *exporter) doExportEvents(
	ctx context.Context,
	requestID string,
	identifier *telemetrygatewayv1.Identifier,
	events []*telemetrygatewayv1.Event,
) ([]string, error) {
	// Start the stream
	stream, err := e.client.RecordEvents(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start export")
	}

	// Send initial metadata
	if err := stream.Send(&telemetrygatewayv1.RecordEventsRequest{
		Payload: &telemetrygatewayv1.RecordEventsRequest_Metadata{
			Metadata: &telemetrygatewayv1.RecordEventsRequestMetadata{
				RequestId:  requestID,
				Identifier: identifier,
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "send initial metadata")
	}

	// Set up a callback that makes sure we pick up all responses from the
	// server.
	collectResults := func() ([]string, error) {
		// We're collecting results now - end the request send stream. From here,
		// the server will eventually get io.EOF and return, then we will eventually
		// get an io.EOF and return. Discard the error because we don't really
		// care - in examples, the error gets discarded as well:
		// https://github.com/grpc/grpc-go/blob/130bc4281c39ac1ed287ec988364d36322d3cd34/examples/route_guide/client/client.go#L145
		//
		// If anything goes wrong stream.Recv() will let us know.
		_ = stream.CloseSend()

		// Wait for responses from server.
		succeededEvents := make([]string, 0, len(events))
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return succeededEvents, err
			}
			if len(resp.GetSucceededEvents()) > 0 {
				succeededEvents = append(succeededEvents, resp.GetSucceededEvents()...)
			}
		}
		if len(succeededEvents) < len(events) {
			return succeededEvents, errors.Newf("%d events did not get recorded successfully",
				len(events)-len(succeededEvents))
		}
		return succeededEvents, nil
	}

	// Start streaming our set of events, chunking them based on message size
	// as determined internally by chunk.Chunker.
	chunker := chunk.New(func(chunkedEvents []*telemetrygatewayv1.Event) error {
		return stream.Send(&telemetrygatewayv1.RecordEventsRequest{
			Payload: &telemetrygatewayv1.RecordEventsRequest_Events{
				Events: &telemetrygatewayv1.RecordEventsRequest_EventsPayload{
					Events: chunkedEvents,
				},
			},
		})
	})
	if err := chunker.Send(events...); err != nil {
		succeeded, _ := collectResults()
		return succeeded, errors.Wrap(err, "chunk and send events")
	}
	if err := chunker.Flush(); err != nil {
		succeeded, _ := collectResults()
		return succeeded, errors.Wrap(err, "flush events")
	}

	return collectResults()
}

func (e *exporter) Close() error { return e.conn.Close() }

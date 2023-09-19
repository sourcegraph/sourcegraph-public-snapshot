package server

import (
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

var meter = otel.GetMeterProvider().Meter("cmd/telemetry-gateway/internal/server")

type Server struct {
	logger      log.Logger
	eventsTopic pubsub.TopicClient

	histogramRecordEventPayloads metric.Int64Histogram

	// Fallback unimplemented handler
	telemetrygatewayv1.UnimplementedTelemeteryGatewayServiceServer
}

var _ telemetrygatewayv1.TelemeteryGatewayServiceServer = (*Server)(nil)

func New(logger log.Logger, eventsTopic pubsub.TopicClient) (*Server, error) {
	recordEventPayloadsHistogram, err := meter.Int64Histogram("telemetry-gateway.record_event_payloads")
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:      logger.Scoped("server", "grpc server"),
		eventsTopic: eventsTopic,

		histogramRecordEventPayloads: recordEventPayloadsHistogram,
	}, nil
}

func (s *Server) RecordEvents(stream telemetrygatewayv1.TelemeteryGatewayService_RecordEventsServer) error {
	var (
		logger = trace.Logger(stream.Context(), s.logger)
		// publisher is initialized once for RecordEventsRequestMetadata.
		publisher *events.Publisher
	)

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		switch msg.Payload.(type) {
		case *telemetrygatewayv1.RecordEventsRequest_Metadata:
			if publisher != nil {
				return status.Error(codes.InvalidArgument, "received metadata more than once")
			}

			metadata := msg.GetMetadata()
			logger = logger.With(log.String("request_id", metadata.GetRequestId()))
			if metadata.GetLicenseKey() == "" {
				return status.Error(codes.InvalidArgument, "metadata missing license key")
			}
			// TODO: Should we include this on events? Should we really validate
			// the key, or just get the metadata?
			licenseInfo, _, err := licensing.ParseProductLicenseKeyWithBuiltinOrGenerationKey(metadata.GetLicenseKey())
			if err != nil {
				return status.Errorf(codes.InvalidArgument, "invalid license key: %s", err)
			}
			logger.Info("handling events submission stream",
				log.Stringp("license.salesforceOpportunityID", licenseInfo.SalesforceOpportunityID),
				log.Stringp("license.salesforceSubscriptionID", licenseInfo.SalesforceSubscriptionID))

			// Set up a publisher with the provided metadata
			publisher, err = events.NewPublisherForStream(s.eventsTopic, metadata)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to create publisher: %v", err)
			}

		case *telemetrygatewayv1.RecordEventsRequest_Events:
			events := msg.GetEvents().GetEvents()
			if len(events) == 0 {
				continue
			}
			if publisher == nil {
				return status.Error(codes.InvalidArgument, "metadata not yet received")
			}

			// Send off our events
			results := publisher.Publish(stream.Context(), events)

			// Aggregate failure details
			message, errFields, succeeded, failed := summarizeResults(results)

			// Record succeeded and failed separately
			s.histogramRecordEventPayloads.Record(stream.Context(), int64(len(succeeded)),
				metric.WithAttributes(attribute.Bool("succeeded", true)))
			s.histogramRecordEventPayloads.Record(stream.Context(), int64(len(failed)),
				metric.WithAttributes(attribute.Bool("succeeded", false)))

			// Generate a log message for diagnostics
			summaryFields := []log.Field{
				log.Int("submitted", len(events)),
				log.Int("succeeded", len(succeeded)),
				log.Int("failed", len(failed)),
			}
			if len(failed) > 0 {
				logger.Error(message, append(summaryFields, errFields...)...)
			} else {
				logger.Info(message, summaryFields...)
			}

			// Let the client know what happened
			if err := stream.Send(&telemetrygatewayv1.RecordEventsResponse{
				SucceededEvents: succeeded,
			}); err != nil {
				return err
			}

		case nil:
			continue

		default:
			return status.Errorf(codes.InvalidArgument, "got malformed message %T", msg.Payload)
		}
	}

	logger.Info("request done")
	return nil
}

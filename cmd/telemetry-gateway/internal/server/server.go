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

		recordingSuccess = make([]*telemetrygatewayv1.RecordEventsResponse_RecordingSuccess, 0)
		recordingErrors  = make([]*telemetrygatewayv1.RecordEventsResponse_RecordingError, 0)
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

			// Simple validation on our most important fields.
			metadata := msg.GetMetadata()
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
				log.Stringp("salesforceOpportunityID", licenseInfo.SalesforceOpportunityID),
				log.Stringp("salesforceSubscriptionID", licenseInfo.SalesforceSubscriptionID))

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

			results := publisher.Publish(stream.Context(), events)

			// Aggregate failure details
			message, errFields, succeeded, failed := summarizeResults(results)

			// Collect results
			if len(succeeded) > 0 {
				recordingSuccess = append(recordingSuccess, succeeded...)
			}
			if len(failed) > 0 {
				recordingErrors = append(recordingErrors, failed...)
			}

			// Generate a log message for diagnostics
			logger.With(errFields...).Error(message,
				log.Int("submitted", len(events)),
				log.Int("succeeded", len(succeeded)),
				log.Int("failed", len(failed)))

			// Record succeeded and failed separately
			s.histogramRecordEventPayloads.Record(stream.Context(), int64(len(succeeded)),
				metric.WithAttributes(attribute.Bool("succeeded", true)))
			s.histogramRecordEventPayloads.Record(stream.Context(), int64(len(failed)),
				metric.WithAttributes(attribute.Bool("succeeded", false)))

		case nil:
			continue

		default:
			return status.Errorf(codes.InvalidArgument, "got malformed message %T", msg.Payload)
		}
	}

	return stream.SendAndClose(&telemetrygatewayv1.RecordEventsResponse{
		SucceededEvents: recordingSuccess,
		FailedEvents:    recordingErrors,
	})
}

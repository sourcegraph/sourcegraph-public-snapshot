package server

import (
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

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
			if failedEvents := results.FailedEvents(); len(failedEvents) == 0 {
				// Record suceeded
				s.histogramRecordEventPayloads.Record(stream.Context(), int64(len(events)),
					metric.WithAttributes(attribute.Bool("succeeded", true)))
			} else {
				// Record succeeded and failed separately
				s.histogramRecordEventPayloads.Record(stream.Context(), int64(len(events)-len(failedEvents)),
					metric.WithAttributes(attribute.Bool("succeeded", true)))
				s.histogramRecordEventPayloads.Record(stream.Context(), int64(len(failedEvents)),
					metric.WithAttributes(attribute.Bool("succeeded", false)))

				// Aggregate failure details
				message, errFields, details := summarizeFailedEvents(len(events), failedEvents)

				// Generate a log message for diagnostics
				logger.With(errFields...).Error(message,
					log.Int("submitted", len(events)),
					log.Int("failed", len(failedEvents)),
					// Monitor if we run into https://github.com/grpc/grpc-go/issues/4265
					log.Int("details.protoSizeBytes", proto.Size(details)))

				// Attach error set to the error response
				st, err := status.New(codes.Internal, message).WithDetails(details)
				if err != nil {
					logger.Error("failed to marshal error status",
						log.Error(err))
					// Just return a failure message to the client without
					// details if we can't marshal the error status
					return status.Error(codes.Internal, message)
				}
				return st.Err()
			}

		case nil:
			continue

		default:
			return status.Errorf(codes.InvalidArgument, "got malformed message %T", msg.Payload)
		}
	}

	return stream.SendAndClose(&telemetrygatewayv1.RecordEventsResponse{})
}

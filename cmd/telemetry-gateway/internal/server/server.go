package server

import (
	"fmt"
	"io"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

type Server struct {
	logger      log.Logger
	eventsTopic pubsub.TopicPublisher

	recordEventsMetrics recordEventsMetrics

	// Fallback unimplemented handler
	telemetrygatewayv1.UnimplementedTelemeteryGatewayServiceServer
}

var _ telemetrygatewayv1.TelemeteryGatewayServiceServer = (*Server)(nil)

func New(logger log.Logger, eventsTopic pubsub.TopicPublisher) (*Server, error) {
	m, err := newRecordEventsMetrics()
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:      logger.Scoped("server"),
		eventsTopic: eventsTopic,

		recordEventsMetrics: m,
	}, nil
}

func (s *Server) RecordEvents(stream telemetrygatewayv1.TelemeteryGatewayService_RecordEventsServer) (err error) {
	var (
		logger = sgtrace.Logger(stream.Context(), s.logger)
		// publisher is initialized once for RecordEventsRequestMetadata.
		publisher *events.Publisher
		// count of all processed events, collected at the end of a request
		totalProcessedEvents int64
	)

	defer func() {
		s.recordEventsMetrics.totalLength.Record(stream.Context(),
			totalProcessedEvents,
			metric.WithAttributes(attribute.Bool("error", err != nil)))
	}()

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

			// Validate self-reported instance identifier
			switch metadata.GetIdentifier().Identifier.(type) {
			case *telemetrygatewayv1.Identifier_LicensedInstance:
				identifier := metadata.Identifier.GetLicensedInstance()
				licenseInfo, _, err := licensing.ParseProductLicenseKey(identifier.GetLicenseKey())
				if err != nil {
					return status.Errorf(codes.InvalidArgument, "invalid license_key: %s", err)
				}
				logger.Info("handling events submission stream for licensed instance",
					log.String("instanceID", identifier.InstanceId),
					log.Stringp("license.salesforceOpportunityID", licenseInfo.SalesforceOpportunityID),
					log.Stringp("license.salesforceSubscriptionID", licenseInfo.SalesforceSubscriptionID))

			case *telemetrygatewayv1.Identifier_UnlicensedInstance:
				identifier := metadata.Identifier.GetUnlicensedInstance()
				if identifier.InstanceId == "" {
					return status.Error(codes.InvalidArgument, "instance_id is required for unlicensed instance")
				}
				logger.Info("handling events submission stream for unlicensed instance",
					log.String("instanceID", identifier.InstanceId))

			default:
				logger.Error("unknown identifier type",
					log.String("type", fmt.Sprintf("%T", metadata.Identifier.Identifier)))
				return status.Error(codes.Unimplemented, "unsupported identifier type")
			}

			// Set up a publisher with the provided metadata
			publisher, err = events.NewPublisherForStream(s.eventsTopic, metadata)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to create publisher: %v", err)
			}

		case *telemetrygatewayv1.RecordEventsRequest_Events:
			events := msg.GetEvents().GetEvents()
			if publisher == nil {
				return status.Error(codes.InvalidArgument, "got events when metadata not yet received")
			}

			// Handle legacy exporters
			migrateEvents(events)

			// Publish events
			resp := handlePublishEvents(
				stream.Context(),
				logger,
				&s.recordEventsMetrics.payload,
				publisher,
				events)

			// Update total count
			totalProcessedEvents += int64(len(events))

			// Let the client know what happened
			if err := stream.Send(resp); err != nil {
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

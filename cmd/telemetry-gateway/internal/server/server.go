package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/server/samsm2m"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

// APIServer is the Telemetry Gateway API server.
type APIServer struct {
	logger      log.Logger
	eventsTopic pubsub.TopicPublisher
	publishOpts events.PublishStreamOptions

	// samsClient is used for M2M authn/authz: go/sams-m2m
	samsClient *sams.ClientV1

	recordEventsMetrics recordEventsMetrics
	recordEventMetrics  recordEventMetrics

	// Fallback unimplemented handler
	telemetrygatewayv1.UnimplementedTelemeteryGatewayServiceServer
}

var _ telemetrygatewayv1.TelemeteryGatewayServiceServer = (*APIServer)(nil)

func New(
	logger log.Logger,
	eventsTopic pubsub.TopicPublisher,
	samsClient *sams.ClientV1,
	publishOpts events.PublishStreamOptions,
) (*APIServer, error) {
	recordEventsRPCMetrics, err := newRecordEventsMetrics()
	if err != nil {
		return nil, err
	}
	recordEventRPCMetrics, err := newRecordEventMetrics()
	if err != nil {
		return nil, err
	}

	return &APIServer{
		logger:      logger.Scoped("api"),
		eventsTopic: eventsTopic,
		publishOpts: publishOpts,

		samsClient: samsClient,

		recordEventsMetrics: recordEventsRPCMetrics,
		recordEventMetrics:  recordEventRPCMetrics,
	}, nil
}

func (s *APIServer) RecordEvents(stream telemetrygatewayv1.TelemeteryGatewayService_RecordEventsServer) (err error) {
	var (
		logger = sgtrace.Logger(stream.Context(), s.logger).Scoped("RecordEvents")
		// publisher is initialized once for RecordEventsRequestMetadata.
		publisher *events.Publisher
		// count of all processed events, collected at the end of a request
		totalProcessedEvents int64
		// use for logging very long-lasting batches
		start = time.Now()
	)

	defer func() {
		s.recordEventsMetrics.totalLength.Record(stream.Context(),
			totalProcessedEvents,
			metric.WithAttributes(
				attribute.Bool("error", err != nil),
				attribute.String("source", publisher.GetSourceName()),
			))
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
			logger = logger.With(log.String("requestID", metadata.GetRequestId()))

			// Validate self-reported instance identifier
			switch metadata.GetIdentifier().GetIdentifier().(type) {
			case *telemetrygatewayv1.Identifier_LicensedInstance:
				identifier := metadata.Identifier.GetLicensedInstance()
				licenseInfo, _, err := licensing.ParseProductLicenseKey(identifier.GetLicenseKey())
				if err != nil {
					return status.Errorf(codes.InvalidArgument, "invalid license_key: %s", err)
				}
				logger = logger.With(log.String("instanceID", identifier.InstanceId))
				// Record start of stream + additional diagnostics details
				// like salesforce info and external URL once
				logger.Info("handling events submission stream for licensed instance",
					log.String("instanceExternalURL", identifier.ExternalUrl),
					log.Stringp("license.salesforceOpportunityID", licenseInfo.SalesforceOpportunityID),
					log.Stringp("license.salesforceSubscriptionID", licenseInfo.SalesforceSubscriptionID))

			case *telemetrygatewayv1.Identifier_UnlicensedInstance:
				identifier := metadata.Identifier.GetUnlicensedInstance()
				if identifier.InstanceId == "" {
					return status.Error(codes.InvalidArgument, "instance_id is required for unlicensed instance")
				}
				logger = logger.With(log.String("instanceID", identifier.InstanceId))
				// Record start of stream
				logger.Info("handling events submission stream for unlicensed instance")

			case *telemetrygatewayv1.Identifier_ManagedService:
				identifier := metadata.Identifier.GetManagedService()
				if identifier.ServiceId == "" {
					return status.Error(codes.InvalidArgument, "service_id is required for managed services")
				}
				logger = logger.With(
					log.String("serviceID", identifier.ServiceId),
					log.Stringp("serviceEnvironment", identifier.ServiceEnvironment))

				// 🚨 SECURITY: Only known clients registered in SAMS can submit events
				// as a managed service.
				if err := samsm2m.CheckWriteEventsScope(stream.Context(), logger, s.samsClient.Tokens()); err != nil {
					return err
				}

				logger.Info("handling events submission stream for managed service")

			default:
				logger.Error("identifier not supported for this RPC",
					log.String("type", fmt.Sprintf("%T", metadata.Identifier.Identifier)))
				return status.Error(codes.Unimplemented, "unsupported identifier type")
			}

			// Set up a publisher with the provided metadata
			publisher, err = events.NewPublisherForStream(logger, s.eventsTopic, metadata, s.publishOpts)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to create publisher: %v", err)
			}
			logger = logger.With(log.String("source", publisher.GetSourceName()))

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

	// Dangerously slow for a single batch, we're at risk of hitting other
	// types of timeouts. Add an error for reporting purposes. We may need to
	// ensure the exporting integration is using smaller batches.
	if duration := time.Since(start); duration > 30*time.Second {
		logger.Error("request done",
			log.Error(errors.New("slow request")),
			log.Duration("duration", duration))
	} else {
		logger.Info("request done")
	}
	return nil
}

func (s *APIServer) RecordEvent(ctx context.Context, req *telemetrygatewayv1.RecordEventRequest) (_ *telemetrygatewayv1.RecordEventResponse, err error) {
	var (
		metadata = req.GetMetadata()
		event    = req.GetEvent()
	)
	if event == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	logger := sgtrace.Logger(ctx, s.logger).
		Scoped("RecordEvent").
		With(
			log.String("requestID", metadata.GetRequestId()),
			// Include more liberal amounts of diagnostics because this RPC
			// currently has a more limited audience
			log.String("eventID", event.GetId()),
			log.String("eventFeature", event.GetFeature()),
			log.String("eventAction", event.GetAction()))

	// We only allow a limited set of identifiers to use this RPC for now, as
	// Sourcegraph instances should only use RecordEvents.
	switch metadata.GetIdentifier().GetIdentifier().(type) {
	case *telemetrygatewayv1.Identifier_ManagedService:
		identifier := metadata.Identifier.GetManagedService()
		if identifier.ServiceId == "" {
			return nil, status.Error(codes.InvalidArgument, "service_id is required for managed services")
		}
		logger = logger.With(
			log.String("serviceID", identifier.ServiceId),
			log.Stringp("serviceEnvironment", identifier.ServiceEnvironment))

		// 🚨 SECURITY: Only known clients registered in SAMS can submit events
		// as a managed service.
		if err := samsm2m.CheckWriteEventsScope(ctx, logger, s.samsClient.Tokens()); err != nil {
			return nil, err
		}

	default:
		logger.Error("identifier not supported for this RPC",
			log.String("type", fmt.Sprintf("%T", metadata.Identifier.Identifier)))
		return nil, status.Error(codes.Unimplemented, "unsupported identifier type")
	}

	// Set up a publisher with the provided metadata
	publisher, err := events.NewPublisherForStream(s.logger, s.eventsTopic, metadata, s.publishOpts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create publisher: %v", err)
	}
	logger = logger.With(log.String("source", publisher.GetSourceName()))

	defer func() {
		s.recordEventMetrics.processedEvents.Add(ctx,
			1, // RPC only accepts 1 event at a time
			metric.WithAttributes(
				attribute.Bool("error", err != nil),
				attribute.String("source", publisher.GetSourceName())))
	}()

	// Submit the single event
	results := publisher.Publish(ctx, []*telemetrygatewayv1.Event{event})
	if len(results) != 1 {
		logger.Error("unexpected result when publishing",
			log.Error(errors.Newf("expected 1 result, got %d", len(results))))
		return nil, status.Errorf(codes.Internal, "unexpected publishing issue")
	}
	if err := results[0].PublishError; err != nil {
		errField := log.Error(err)
		code := codes.Internal
		if errors.IsContextCanceled(err) {
			// Don't record an error for context canceled errors, since it's
			// generally not very interesting - the client probably decided to
			// cancel or time out. Instead, just record the error string, so
			// that it doesn't go to Sentry.
			errField = log.String("error", err.Error())
			code = codes.Canceled
		}
		logger.Error("failed to publish event", errField)
		return nil, status.Errorf(code, "failed to publish event: %v", err)
	}

	return &telemetrygatewayv1.RecordEventResponse{
		// no properties
	}, nil
}

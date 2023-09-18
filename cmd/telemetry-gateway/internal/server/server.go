package server

import (
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

type Server struct {
	logger      log.Logger
	eventsTopic pubsub.TopicClient

	// Fallback unimplemented handler
	telemetrygatewayv1.UnimplementedTelemeteryGatewayServiceServer
}

var _ telemetrygatewayv1.TelemeteryGatewayServiceServer = (*Server)(nil)

func New(logger log.Logger, eventsTopic pubsub.TopicClient) *Server {
	return &Server{
		logger:      logger.Scoped("server", "grpc server"),
		eventsTopic: eventsTopic,
	}
}

func (s *Server) RecordEvents(stream telemetrygatewayv1.TelemeteryGatewayService_RecordEventsServer) error {
	var (
		logger = trace.Logger(stream.Context(), s.logger)
		// publisher is initialized once for RecordEventsRequestMetadata.
		publisher *events.Publisher
		// submittedEvents is a count of events submitted so far in the stream.
		submittedEvents int
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
			if failedEvents := results.FailedEvents(); len(failedEvents) > 0 {
				var message string
				if len(failedEvents) == submittedEvents {
					message = "all events failed to submit"
				} else {
					message = "some events failed to submit"
				}

				// Collect details about the events that failed to submit.
				failedEventsDetails := make([]*telemetrygatewayv1.RecordEventsErrorDetails_EventError, len(failedEvents))
				errFields := make([]log.Field, len(failedEvents))
				for i, failure := range failedEvents {
					failedEventsDetails[i] = &telemetrygatewayv1.RecordEventsErrorDetails_EventError{
						EventId: failure.EventID,
						Error:   failure.PublishError.Error(),
					}
					errFields[i] = log.NamedError(fmt.Sprintf("error.%d", i), failure.PublishError)
				}
				details := &telemetrygatewayv1.RecordEventsErrorDetails{
					FailedEvents: failedEventsDetails,
				}

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

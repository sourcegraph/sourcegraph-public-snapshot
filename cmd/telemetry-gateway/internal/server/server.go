package server

import (
	"github.com/sourcegraph/sourcegraph/internal/pubsub"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

type Server struct {
	EventsTopic pubsub.TopicClient

	// Fallback unimplemented handler
	telemetrygatewayv1.UnimplementedTelemeteryGatewayServiceServer
}

var _ telemetrygatewayv1.TelemeteryGatewayServiceServer = (*Server)(nil)

func (s *Server) RecordEvents(stream telemetrygatewayv1.TelemeteryGatewayService_RecordEventsServer) error {
	return nil
}

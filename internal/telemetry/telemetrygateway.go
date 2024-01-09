package telemetry

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// newTelemetryGatewayEvent translates recording to raw events for storage and
// export. It extracts actor from context as the event user.
func newTelemetryGatewayEvent(
	ctx context.Context,
	now time.Time,
	newUUID func() string,
	feature eventFeature,
	action eventAction,
	parameters *EventParameters,
) *telemetrygatewayv1.Event {
	// Assign zero value for ease of reference, and in the proto spec, parameters
	// is not optional.
	if parameters == nil {
		parameters = &EventParameters{}
	}

	event := telemetrygatewayv1.NewEventWithDefaults(ctx, now, newUUID)
	event.Feature = string(feature)
	event.Action = string(action)
	event.Source = &telemetrygatewayv1.EventSource{
		Server: &telemetrygatewayv1.EventSource_Server{
			Version: version.Version(),
		},
		Client: nil, // no client, this is recorded directly in backend
	}
	event.Parameters = &telemetrygatewayv1.EventParameters{
		Version: int32(parameters.Version),
		Metadata: func() map[string]float64 {
			if len(parameters.Metadata) == 0 {
				return nil
			}
			m := make(map[string]float64, len(parameters.Metadata))
			for k, v := range parameters.Metadata {
				m[string(k)] = v
			}
			return m
		}(),
		PrivateMetadata: func() *structpb.Struct {
			if len(parameters.PrivateMetadata) == 0 {
				return nil
			}
			s, err := structpb.NewStruct(parameters.PrivateMetadata)
			if err != nil {
				return &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"telemetry.error": structpb.NewStringValue("failed to marshal private metadata: " + err.Error()),
					},
				}
			}
			return s
		}(),
		BillingMetadata: func() *telemetrygatewayv1.EventBillingMetadata {
			if parameters.BillingMetadata == nil {
				return nil
			}
			return &telemetrygatewayv1.EventBillingMetadata{
				Product:  string(parameters.BillingMetadata.Product),
				Category: string(parameters.BillingMetadata.Category),
			}
		}(),
	}
	return event
}

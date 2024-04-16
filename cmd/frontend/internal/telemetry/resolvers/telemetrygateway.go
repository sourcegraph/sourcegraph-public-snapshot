package resolvers

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newTelemetryGatewayEvents converts GraphQL telemetry input to the Telemetry
// Gateway wire format.
func newTelemetryGatewayEvents(
	ctx context.Context,
	now time.Time,
	newUUID func() string,
	gqlEvents []graphqlbackend.TelemetryEventInput,
) ([]*telemetrygatewayv1.Event, error) {
	gatewayEvents := make([]*telemetrygatewayv1.Event, len(gqlEvents))
	for i, gqlEvent := range gqlEvents {
		event := telemetrygatewayv1.NewEventWithDefaults(ctx, now, newUUID)

		if gqlEvent.Timestamp != nil {
			event.Timestamp = timestamppb.New(gqlEvent.Timestamp.Time)
		}

		event.Feature = gqlEvent.Feature
		event.Action = gqlEvent.Action

		// Override interaction ID, or just set it, if an interaction ID is
		// explicitly provided as part of event data.
		if gqlEvent.Parameters.InteractionID != nil && len(*gqlEvent.Parameters.InteractionID) > 0 {
			if event.Interaction == nil {
				event.Interaction = &telemetrygatewayv1.EventInteraction{}
			}
			event.Interaction.InteractionId = gqlEvent.Parameters.InteractionID
		}

		// Parse metadata
		var metadata map[string]float64
		if gqlEvent.Parameters.Metadata != nil && len(*gqlEvent.Parameters.Metadata) != 0 {
			metadata = make(map[string]float64, len(*gqlEvent.Parameters.Metadata))
			for _, kv := range *gqlEvent.Parameters.Metadata {
				switch v := kv.Value.Value.(type) {
				case int:
					metadata[kv.Key] = float64(v)
				case int8:
					metadata[kv.Key] = float64(v)
				case int32:
					metadata[kv.Key] = float64(v)
				case int64:
					metadata[kv.Key] = float64(v)
				case float32:
					metadata[kv.Key] = float64(v)
				case float64:
					metadata[kv.Key] = v
				default:
					return gatewayEvents,
						errors.Newf("metadata %q has invalid value type %T", kv.Key, v)
				}
			}
		}

		// Parse private metadata
		var privateMetadata *structpb.Struct
		if gqlEvent.Parameters.PrivateMetadata != nil {
			switch v := gqlEvent.Parameters.PrivateMetadata.Value.(type) {
			// If the input is an object, turn it into proto struct as-is
			case map[string]any:
				var err error
				privateMetadata, err = structpb.NewStruct(v)
				if err != nil {
					return nil, errors.Wrapf(err, "error converting privateMetadata to protobuf struct for event %d", i)
				}

			// Otherwise, nest the value within a proto struct
			default:
				protoValue, err := structpb.NewValue(v)
				if err != nil {
					return nil, errors.Wrapf(err, "error converting privateMetadata to protobuf value for event %d", i)
				}
				privateMetadata = &structpb.Struct{
					Fields: map[string]*structpb.Value{"value": protoValue},
				}
			}
		}

		// Configure parameters
		event.Parameters = &telemetrygatewayv1.EventParameters{
			Version:         gqlEvent.Parameters.Version,
			Metadata:        metadata,
			PrivateMetadata: privateMetadata,
			BillingMetadata: func() *telemetrygatewayv1.EventBillingMetadata {
				if gqlEvent.Parameters.BillingMetadata == nil {
					return nil
				}
				return &telemetrygatewayv1.EventBillingMetadata{
					Product:  gqlEvent.Parameters.BillingMetadata.Product,
					Category: gqlEvent.Parameters.BillingMetadata.Category,
				}
			}(),
		}
		event.Source = &telemetrygatewayv1.EventSource{
			Server: &telemetrygatewayv1.EventSource_Server{
				Version: version.Version(),
			},
			Client: &telemetrygatewayv1.EventSource_Client{
				Name:    gqlEvent.Source.Client,
				Version: gqlEvent.Source.ClientVersion,
			},
		}

		if gqlEvent.MarketingTracking != nil {
			event.MarketingTracking = &telemetrygatewayv1.EventMarketingTracking{
				Url:             gqlEvent.MarketingTracking.Url,
				FirstSourceUrl:  gqlEvent.MarketingTracking.FirstSourceURL,
				CohortId:        gqlEvent.MarketingTracking.CohortID,
				Referrer:        gqlEvent.MarketingTracking.Referrer,
				LastSourceUrl:   gqlEvent.MarketingTracking.LastSourceURL,
				DeviceSessionId: gqlEvent.MarketingTracking.DeviceSessionID,
				SessionReferrer: gqlEvent.MarketingTracking.SessionReferrer,
				SessionFirstUrl: gqlEvent.MarketingTracking.SessionFirstURL,
			}
		}

		// Done!
		gatewayEvents[i] = event
	}
	return gatewayEvents, nil
}

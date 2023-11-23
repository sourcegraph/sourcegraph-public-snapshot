package resolvers

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newTelemetryGatewayEvents(
	ctx context.Context,
	now time.Time,
	newUUID func() string,
	gqlEvents []graphqlbackend.TelemetryEventInput,
) ([]*telemetrygatewayv1.Event, error) {
	gatewayEvents := make([]*telemetrygatewayv1.Event, len(gqlEvents))
	for i, gqlEvent := range gqlEvents {
		event := telemetrygatewayv1.NewEventWithDefaults(ctx, now, newUUID)

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
			Version: gqlEvent.Parameters.Version,
			Metadata: func() map[string]int64 {
				if gqlEvent.Parameters.Metadata == nil || len(*gqlEvent.Parameters.Metadata) == 0 {
					return nil
				}
				metadata := make(map[string]int64, len(*gqlEvent.Parameters.Metadata))
				for _, kv := range *gqlEvent.Parameters.Metadata {
					metadata[kv.Key] = int64(kv.Value)
				}
				return metadata
			}(),
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

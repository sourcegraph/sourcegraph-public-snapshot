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
	events []graphqlbackend.TelemetryEventInput,
) ([]*telemetrygatewayv1.Event, error) {
	gatewayEvents := make([]*telemetrygatewayv1.Event, len(events))
	for i, e := range events {
		event := telemetrygatewayv1.NewEventWithDefaults(ctx, now, newUUID)

		event.Feature = e.Feature
		event.Action = e.Action

		// Override the interaction ID if one is explicitly provided
		if e.Parameters.InteractionID != nil && len(*e.Parameters.InteractionID) > 0 {
			if event.Interaction == nil {
				event.Interaction = &telemetrygatewayv1.EventInteraction{}
			}
		}

		// Parse private metadata
		var privateMetadata *structpb.Struct
		if e.Parameters.PrivateMetadata != nil {
			switch v := e.Parameters.PrivateMetadata.Value.(type) {
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
			Version: e.Parameters.Version,
			Metadata: func() map[string]int64 {
				if e.Parameters.Metadata == nil || len(*e.Parameters.Metadata) == 0 {
					return nil
				}
				metadata := make(map[string]int64, len(*e.Parameters.Metadata))
				for _, kv := range *e.Parameters.Metadata {
					metadata[kv.Key] = int64(kv.Value)
				}
				return metadata
			}(),
			PrivateMetadata: privateMetadata,
			BillingMetadata: func() *telemetrygatewayv1.EventBillingMetadata {
				if e.Parameters.BillingMetadata == nil {
					return nil
				}
				return &telemetrygatewayv1.EventBillingMetadata{
					Product:  e.Parameters.BillingMetadata.Product,
					Category: e.Parameters.BillingMetadata.Category,
				}
			}(),
		}
		event.Source = &telemetrygatewayv1.EventSource{
			Server: &telemetrygatewayv1.EventSource_Server{
				Version: version.Version(),
			},
			Client: &telemetrygatewayv1.EventSource_Client{
				Name:    e.Source.Client,
				Version: e.Source.ClientVersion,
			},
		}

		if e.MarketingTracking != nil {
			event.MarketingTracking = &telemetrygatewayv1.EventMarketingTracking{
				Url:             e.MarketingTracking.Url,
				FirstSourceUrl:  e.MarketingTracking.FirstSourceURL,
				CohortId:        e.MarketingTracking.CohortID,
				Referrer:        e.MarketingTracking.Referrer,
				LastSourceUrl:   e.MarketingTracking.LastSourceURL,
				DeviceSessionId: e.MarketingTracking.DeviceSessionID,
				SessionReferrer: e.MarketingTracking.SessionReferrer,
				SessionFirstUrl: e.MarketingTracking.SessionFirstURL,
			}
		}

		// Done!
		gatewayEvents[i] = event
	}
	return gatewayEvents, nil
}

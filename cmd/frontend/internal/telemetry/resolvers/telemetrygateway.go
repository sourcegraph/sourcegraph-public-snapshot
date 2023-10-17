package resolvers

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RecordEvent(ctx context.Context, args *graphqlbackend.RecordEventArgs) (*graphqlbackend.EmptyResponse, error) {

	// Create telemetry events
	events, err := createTelemetryEvents(ctx, args)
	if err != nil {
		return nil, err
	}

	// Record gateway events
	err = recordGatewayEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}
func createTelemetryEvents(ctx context.Context, args *graphqlbackend.RecordEventArgs) ([]*telemetrygatewayv1.Event, error) {
	// func RecordEvent(ctx context.Context, args *graphqlbackend.RecordEventArgs) (*graphqlbackend.EmptyResponse, error) {
	var metadata []graphqlbackend.TelemetryEventMetadataInput
	if args.Event.Parameters.Metadata != nil {
		for _, m := range *args.Event.Parameters.Metadata {
			metadata = append(metadata, graphqlbackend.TelemetryEventMetadataInput{
				Key:   m.Key,
				Value: m.Value,
			})
		}
	}

	// Translate args to events
	events := []graphqlbackend.TelemetryEventInput{
		{
			Feature: args.Event.Feature,
			Action:  args.Event.Action,
			Source: graphqlbackend.TelemetryEventSourceInput{
				Client:        args.Event.Source.Client,
				ClientVersion: args.Event.Source.ClientVersion,
			},
			Parameters: &graphqlbackend.TelemetryEventParametersInput{
				Version:         args.Event.Parameters.Version,
				Metadata:        &metadata,
				PrivateMetadata: args.Event.Parameters.PrivateMetadata,
				BillingMetadata: &graphqlbackend.TelemetryEventBillingMetadataInput{
					Product:  args.Event.Parameters.BillingMetadata.Product,
					Category: args.Event.Parameters.BillingMetadata.Category,
				},
			},
			MarketingTracking: &graphqlbackend.TelemetryEventMarketingTrackingInput{
				Url:             args.Event.MarketingTracking.Url,
				FirstSourceURL:  args.Event.MarketingTracking.FirstSourceURL,
				CohortID:        args.Event.MarketingTracking.CohortID,
				Referrer:        args.Event.MarketingTracking.Referrer,
				LastSourceURL:   args.Event.MarketingTracking.LastSourceURL,
				DeviceSessionID: args.Event.MarketingTracking.DeviceSessionID,
				SessionReferrer: args.Event.MarketingTracking.SessionReferrer,
				SessionFirstURL: args.Event.MarketingTracking.SessionFirstURL,
			},
		},
	}

	// Call newTelemetryGatewayEvents
	gatewayEvents, err := newTelemetryGatewayEvents(ctx, time.Now(), uuid.NewString, events)
	if err != nil {
		return nil, err
	}

	return gatewayEvents, nil
}

func recordGatewayEvents(ctx context.Context, events []*telemetrygatewayv1.Event) error {

	// Store events

	// Export events

	return nil
}

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

		// Parse private metadata
		var privateMetadata *structpb.Struct
		if e.Parameters.PrivateMetadata != nil && len(*e.Parameters.PrivateMetadata) > 0 {
			data, err := e.Parameters.PrivateMetadata.MarshalJSON()
			if err != nil {
				return nil, errors.Wrapf(err, "error marshaling privateMetadata for event %d", i)
			}
			var privateData map[string]any
			if err := json.Unmarshal(data, &privateData); err != nil {
				return nil, errors.Wrapf(err, "error unmarshaling privateMetadata for event %d", i)
			}
			privateMetadata, err = structpb.NewStruct(privateData)
			if err != nil {
				return nil, errors.Wrapf(err, "error converting privateMetadata to protobuf for event %d", i)
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

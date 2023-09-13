// Package telemetry implements "Telemetry V2", which supercedes event_logs
// as the mechanism for reporting telemetry from all Sourcegraph instances to
// Sourcergraph.
package telemetry

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/telemetrygateway"
	"github.com/sourcegraph/sourcegraph/internal/version"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

type Event struct {
	Feature    eventFeature
	Action     eventAction
	Parameters EventParameters
}

// constString effectively requires strings to be statically defined constants.
type constString string

// EventMetadata is secure, PII-free metadata that can be attached to events.
// Keys must be const strings.
type EventMetadata map[constString]int64

// EventBillingMetadata records metadata that attributes the event to product
// billing categories.
type EventBillingMetadata struct {
	// Product identifier.
	Product billingProduct
	// Category identifier.
	Category billingCategory
}

type EventParameters struct {
	// Version can be used to denote the "shape" of this event.
	Version int
	// Metadata is PII-free metadata about the event that we export.
	Metadata EventMetadata
	// PrivateMetadata is arbitrary metadata that is generally not exported.
	PrivateMetadata map[string]any
	// BillingMetadata contains metadata we can use for billing purposes.
	BillingMetadata *EventBillingMetadata
}

// EventRecorder is for creating and recording telemetry events in the backend
// using Telemetry V2, which exports events to Sourcergraph.
type EventRecorder struct {
	exporter telemetrygateway.Exporter
}

// ❗ Experimental - do not use!
func NewEventRecorder(exporter telemetrygateway.Exporter) *EventRecorder {
	return &EventRecorder{exporter: exporter}
}

// Record records a single telemetry event with the context's Sourcegraph
// actor.
func (r *EventRecorder) Record(ctx context.Context, feature eventFeature, action eventAction, parameters EventParameters) error {
	return r.exporter.ExportEvents(ctx, []*telemetrygatewayv1.Event{
		makeRawEvent(ctx, time.Now(), feature, action, parameters),
	})
}

// BatchRecord records a set of telemetry events with the context's
// Sourcegraph actor.
func (r *EventRecorder) BatchRecord(ctx context.Context, events ...Event) error {
	if len(events) == 0 {
		return nil
	}
	rawEvents := make([]*telemetrygatewayv1.Event, len(events))
	for i, e := range events {
		rawEvents[i] = makeRawEvent(ctx, time.Now(), e.Feature, e.Action, e.Parameters)
	}
	return r.exporter.ExportEvents(ctx, rawEvents)
}

// makeRawEvent translates recording to raw events for storage and export. It
// extracts actor from context as the event user.
func makeRawEvent(ctx context.Context, now time.Time, feature eventFeature, action eventAction, parameters EventParameters) *telemetrygatewayv1.Event {
	return &telemetrygatewayv1.Event{
		Timestamp: timestamppb.New(now),
		Feature:   string(feature),
		Action:    string(action),
		Source: &telemetrygatewayv1.EventSource{
			Server: &telemetrygatewayv1.EventSource_Server{
				Version: version.Version(),
			},
			Client: nil, // no client, event is being created server-side
		},
		Parameters: &telemetrygatewayv1.EventParameters{
			Version: int32(parameters.Version),
			Metadata: func() map[string]int64 {
				if len(parameters.Metadata) == 0 {
					return nil
				}
				m := make(map[string]int64, len(parameters.Metadata))
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
		},
		User: func() *telemetrygatewayv1.EventUser {
			act := actor.FromContext(ctx)
			if !act.IsAuthenticated() {
				return nil
			}
			return &telemetrygatewayv1.EventUser{
				UserId:          int64(act.UID),
				AnonymousUserId: act.AnonymousUID,
			}
		}(),
	}
}

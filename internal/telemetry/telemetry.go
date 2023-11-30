// Package telemetry implements "Telemetry V2", which supercedes event_logs
// as the mechanism for reporting telemetry from all Sourcegraph instances to
// Sourcergraph.
package telemetry

import (
	"context"
	"time"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

// constString effectively requires strings to be statically defined constants.
type constString string

// EventMetadata is secure, PII-free metadata that can be attached to events.
// Keys must be const strings.
type EventMetadata map[constString]float64

// MetadataBool returns 1 for true and 0 for false, for use in EventMetadata's
// restricted int64 values.
func MetadataBool(value bool) float64 {
	if value {
		return 1 // true
	}
	return 0 // 0
}

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

// EventsStore backs event recorders with storage of events for export.
// In general, prefer to use the telemetryrecorder.New() constructor for a
// default implementation.
//
// In tests, you can use the telemetrytest.NewMockEventsStore() implementation.
type EventsStore interface {
	StoreEvents(context.Context, []*telemetrygatewayv1.Event) error
}

// EventRecorder is for creating and recording telemetry events in the backend
// using Telemetry V2, which exports events to Sourcergraph.
type EventRecorder struct{ store EventsStore }

// NewEventRecorder creates a custom event recorder backed by a store
// implementation. In general, prefer to use the telemetryrecorder.New()
// constructor instead.
//
// In tests, you may want to use this constructor alongside the
// the telemetrytest.NewMockEventsStore() implementation of EventsStore.
func NewEventRecorder(store EventsStore) *EventRecorder {
	return &EventRecorder{store: store}
}

// Record records a single telemetry event with the context's Sourcegraph
// actor. Parameters are optional.
func (r *EventRecorder) Record(ctx context.Context, feature eventFeature, action eventAction, parameters *EventParameters) error {
	return r.store.StoreEvents(ctx, []*telemetrygatewayv1.Event{
		newTelemetryGatewayEvent(ctx, time.Now(), telemetrygatewayv1.DefaultEventIDFunc, feature, action, parameters),
	})
}

type Event struct {
	// Feature is required.
	Feature eventFeature
	// Action is required.
	Action eventAction
	// Parameters are optional.
	Parameters EventParameters
}

// BatchRecord records a set of telemetry events with the context's
// Sourcegraph actor.
func (r *EventRecorder) BatchRecord(ctx context.Context, events ...Event) error {
	if len(events) == 0 {
		return nil
	}
	rawEvents := make([]*telemetrygatewayv1.Event, len(events))
	for i, e := range events {
		rawEvents[i] = newTelemetryGatewayEvent(ctx, time.Now(), telemetrygatewayv1.DefaultEventIDFunc, e.Feature, e.Action, &e.Parameters)
	}
	return r.store.StoreEvents(ctx, rawEvents)
}

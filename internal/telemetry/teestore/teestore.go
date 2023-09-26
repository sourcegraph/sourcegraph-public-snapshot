package teestore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/sensitivemetadataallowlist"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Store tees events into both the event_logs table and the new telemetry export
// queue, translating the message into the existing event_logs format on a
// best-effort basis.
type Store struct {
	exportQueue database.TelemetryEventsExportQueueStore
	eventLogs   database.EventLogStore
}

func NewStore(exportQueue database.TelemetryEventsExportQueueStore, eventLogs database.EventLogStore) *Store {
	return &Store{exportQueue, eventLogs}
}

func (s *Store) StoreEvents(ctx context.Context, events []*telemetrygatewayv1.Event) error {
	// Write to both stores at the same time.
	wg := pool.New().WithErrors()
	wg.Go(func() error {
		if err := s.exportQueue.QueueForExport(ctx, events); err != nil {
			return errors.Wrap(err, "bulk inserting telemetry events")
		}
		return nil
	})
	if !shouldDisableV1(ctx) {
		wg.Go(func() error {
			if err := s.eventLogs.BulkInsert(ctx, toEventLogs(time.Now, events)); err != nil {
				return errors.Wrap(err, "bulk inserting events logs")
			}
			return nil
		})
	}
	return wg.Wait()
}

func toEventLogs(now func() time.Time, telemetryEvents []*telemetrygatewayv1.Event) []*database.Event {
	sensitiveMetadataAllowlist := sensitivemetadataallowlist.AllowedEventTypes()

	eventLogs := make([]*database.Event, len(telemetryEvents))
	for i, e := range telemetryEvents {
		// Note that all generated proto getters are nil-safe, so use those to
		// get fields rather than accessing fields directly.
		eventLogs[i] = &database.Event{
			ID:       0,   // not required on insert
			InsertID: nil, // not required on insert

			// Identifiers
			Name: fmt.Sprintf("%s.%s", e.GetFeature(), e.GetAction()),
			Timestamp: func() time.Time {
				if e.GetTimestamp() == nil {
					return now()
				}
				return e.GetTimestamp().AsTime()
			}(),

			// User
			UserID:          uint32(e.GetUser().GetUserId()),
			AnonymousUserID: e.GetUser().GetAnonymousUserId(),

			// GetParameters.Metadata
			PublicArgument: func() json.RawMessage {
				md := e.GetParameters().GetMetadata()
				mdPayload := make(map[string]any, len(md))
				for k, v := range md {
					mdPayload[k] = v
				}
				// Attach a simple indicator to denote if this event will
				// be exported.
				mdPayload["telemetry.event.exportable"] = true

				data, err := json.Marshal(mdPayload)
				if err != nil {
					data, _ = json.Marshal(map[string]string{"marshal.error": err.Error()})
				}
				return data
			}(),

			// GetParameters.PrivateMetadata
			Argument: func() json.RawMessage {
				md := e.GetParameters().GetPrivateMetadata().AsMap()
				if len(md) == 0 {
					return nil
				}

				// Attach a simple indicator to denote if this metadata will
				// be exported.
				md["telemetry.privateMetadata.exportable"] = sensitiveMetadataAllowlist.IsAllowed(e)

				data, err := json.Marshal(md)
				if err != nil {
					data, _ = json.Marshal(map[string]string{"marshal.error": err.Error()})
				}
				return data
			}(),

			// Parameters.BillingMetadata
			BillingProductCategory: pointers.NonZeroPtr(e.GetParameters().GetBillingMetadata().GetCategory()),
			BillingEventID:         nil, // No equivalents in telemetry events

			// Source.Client
			Source: func() string {
				if source := e.GetSource().GetClient().GetName(); source != "" {
					return source
				}
				return "BACKEND" // must be non-empty
			}(),
			Client: func() *string {
				if c := e.GetSource().GetClient(); c != nil {
					return pointers.Ptr(fmt.Sprintf("%s:%s",
						c.GetName(), c.GetVersion()))
				}
				return nil
			}(),

			// Source.Server
			Version: e.GetSource().GetServer().GetVersion(),

			// MarketingTracking
			URL:            e.GetMarketingTracking().GetUrl(),
			CohortID:       pointers.NonZeroPtr(e.GetMarketingTracking().GetCohortId()),
			FirstSourceURL: pointers.NonZeroPtr(e.GetMarketingTracking().GetFirstSourceUrl()),
			LastSourceURL:  pointers.NonZeroPtr(e.GetMarketingTracking().GetLastSourceUrl()),
			Referrer:       pointers.NonZeroPtr(e.GetMarketingTracking().GetReferrer()),
			DeviceID:       pointers.NonZeroPtr(e.GetMarketingTracking().GetDeviceSessionId()),

			// FeatureFlags
			EvaluatedFlagSet: func() featureflag.EvaluatedFlagSet {
				flags := e.GetFeatureFlags().GetFlags()
				set := make(featureflag.EvaluatedFlagSet, len(flags))
				for k, v := range flags {
					// We can expect all values to be bools for now
					set[k], _ = strconv.ParseBool(v)
				}
				return set
			}(),
		}
	}
	return eventLogs
}

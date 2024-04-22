package telemetrygatewayevent

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/requestinteraction"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

// DefaultEventIDFunc is the default generator for telemetry event IDs.
// We currently use V7, which is time-ordered, making them useful for event IDs.
// https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format-03#name-uuid-version-7
var DefaultEventIDFunc = telemetrygatewayv1.DefaultEventIDFunc

// New creates a uniform event with defaults filled in, including any relevant
// data required from context. All constructors making raw events for Sourcegraph
// instances to export MUST start with this.
func New(ctx context.Context, now time.Time, newEventID func() string) *telemetrygatewayv1.Event {
	return &telemetrygatewayv1.Event{
		Id:        newEventID(),
		Timestamp: timestamppb.New(now),
		Interaction: func() *telemetrygatewayv1.EventInteraction {
			// Trace associated with event is the same trace on the event recording
			// request where the event is being created, as they should all happen
			// within the interaction, even when recording a set of events e.g. from
			// buffering.
			var traceID *string
			if eventTrace := trace.FromContext(ctx).SpanContext(); eventTrace.IsValid() {
				traceID = pointers.Ptr(eventTrace.TraceID().String())
			}

			// Get the interaction ID if provided
			var interactionID *string
			if it := requestinteraction.FromContext(ctx); it != nil {
				interactionID = pointers.Ptr(it.ID)
			}

			// Get geolocation of request client, if there is one.
			var geolocation *telemetrygatewayv1.EventInteraction_Geolocation
			if rc := requestclient.FromContext(ctx); rc != nil {
				if cc, err := rc.OriginCountryCode(); err == nil {
					geolocation = &telemetrygatewayv1.EventInteraction_Geolocation{
						CountryCode: cc,
					}
				}
			}

			// If we have nothing interesting to show, leave out Interaction
			// entirely.
			if traceID == nil && interactionID == nil && geolocation == nil {
				return nil
			}

			return &telemetrygatewayv1.EventInteraction{
				TraceId:       traceID,
				InteractionId: interactionID,
				Geolocation:   geolocation,
			}
		}(),
		User: func() *telemetrygatewayv1.EventUser {
			act := actor.FromContext(ctx)
			if !act.IsAuthenticated() && act.AnonymousUID == "" {
				return nil
			}
			return &telemetrygatewayv1.EventUser{
				UserId:          pointers.NonZeroPtr(int64(act.UID)),
				AnonymousUserId: pointers.NonZeroPtr(act.AnonymousUID),
			}
		}(),
		FeatureFlags: func() *telemetrygatewayv1.EventFeatureFlags {
			flags := featureflag.GetEvaluatedFlagSet(ctx)
			if len(flags) == 0 {
				return nil
			}
			data := make(map[string]string, len(flags))
			for k, v := range flags {
				data[k] = strconv.FormatBool(v)
			}
			return &telemetrygatewayv1.EventFeatureFlags{
				Flags: data,
			}
		}(),
	}
}

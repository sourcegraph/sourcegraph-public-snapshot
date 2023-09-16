package v1

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// DefaultEventIDFunc is the default generator for telemetry event IDs.
var DefaultEventIDFunc = uuid.NewString

// NewEventWithDefaults creates a uniform event with defaults filled in. All
// constructors making raw events should start with this. In particular, this
// adds any relevant data required from context.
func NewEventWithDefaults(ctx context.Context, now time.Time, newEventID func() string) *Event {
	return &Event{
		Id:        newEventID(),
		Timestamp: timestamppb.New(now),
		User: func() *EventUser {
			act := actor.FromContext(ctx)
			if !act.IsAuthenticated() && act.AnonymousUID == "" {
				return nil
			}
			return &EventUser{
				UserId:          pointers.NonZeroPtr(int64(act.UID)),
				AnonymousUserId: pointers.NonZeroPtr(act.AnonymousUID),
			}
		}(),
		FeatureFlags: func() *EventFeatureFlags {
			flags := featureflag.GetEvaluatedFlagSet(ctx)
			if len(flags) == 0 {
				return nil
			}
			data := make(map[string]string, len(flags))
			for k, v := range flags {
				data[k] = strconv.FormatBool(v)
			}
			return &EventFeatureFlags{
				Flags: data,
			}
		}(),
	}
}

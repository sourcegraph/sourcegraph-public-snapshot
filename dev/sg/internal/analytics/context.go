package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/dev/okay"
)

type analyticsStoreKey struct{}

// WithContext enables analytics in this context.
func WithContext(ctx context.Context, sgVersion string) context.Context {
	return context.WithValue(ctx, analyticsStoreKey{}, &eventStore{
		sgVersion: sgVersion,
		events:    make([]*okay.Event, 0, 10),
	})
}

// LogDuration tracks an event in the per-run analytics store, if analytics are enabled,
// in the context of a command.
func LogDuration(ctx context.Context, name string, labels []string, duration time.Duration) {
	store, ok := ctx.Value(analyticsStoreKey{}).(*eventStore)
	if !ok {
		return
	}
	store.events = append(store.events, &okay.Event{
		Name:      name,
		Labels:    labels,
		Timestamp: time.Now().Add(-duration), // Timestamp as start of event
		Metrics: map[string]okay.Metric{
			"duration": okay.Duration(duration),
		},
		UniqueKey: []string{"event_id"},
		Properties: map[string]string{
			"event_id": uuid.NewString(),
		},
	})
}

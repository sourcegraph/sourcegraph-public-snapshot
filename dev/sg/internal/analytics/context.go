package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/dev/okay"
)

type eventStoreKey struct{}

// WithContext enables analytics in this context.
func WithContext(ctx context.Context, sgVersion string) context.Context {
	return context.WithValue(ctx, eventStoreKey{}, &eventStore{
		sgVersion: sgVersion,
		events:    make([]*okay.Event, 0, 10),
	})
}

// getStore retrieves the events store from context if it exists. Callers should check
// that the store is non-nil before attempting to use it.
func getStore(ctx context.Context) *eventStore {
	store, ok := ctx.Value(eventStoreKey{}).(*eventStore)
	if !ok {
		return nil
	}
	return store
}

// LogEvent tracks an event in the per-run analytics store, if analytics are enabled,
// in the context of a command.
//
// Events can also be provided to indicate that something happened - for example, and
// error or cancellation. These are treated as metrics with a count of 1.
func LogEvent(ctx context.Context, name string, labels []string, startedAt time.Time, events ...string) {
	store := getStore(ctx)
	if store == nil {
		return
	}

	metrics := map[string]okay.Metric{
		"duration": okay.Duration(time.Since(startedAt)),
	}
	for _, event := range events {
		metrics[event] = okay.Count(1)
	}

	store.events = append(store.events, &okay.Event{
		Name:      name,
		Labels:    labels,
		Timestamp: startedAt, // Timestamp as start of event
		Metrics:   metrics,
		UniqueKey: []string{"event_id"},
		Properties: map[string]string{
			"event_id": uuid.NewString(),
		},
	})
}

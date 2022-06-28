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
// In general, usage should be as follows:
//
// - category denotes the category of the event, such as "lint_runner:.
// - labels denote subcategories this event belongs to, such as the specific lint runner.
// - events denote what happened as part of this logged event, such as "failed" or
//   "succeeded". These are treated as metrics with a count of 1.
//
// Events are automatically created with a duration relative to the provided start time,
// and persisted to disk at the end of command execution.
//
// It returns the event that was created so that you can add additional metadata if
// desired - use sparingly.
func LogEvent(ctx context.Context, category string, labels []string, startedAt time.Time, events ...string) *okay.Event {
	// Validate the incoming event
	if category == "" {
		panic("LogEvent.category must be set")
	}
	if startedAt.IsZero() {
		panic("LogEvent.startedAt must be a valid time")
	}

	// Set events as metrics
	metrics := map[string]okay.Metric{
		"duration": okay.Duration(time.Since(startedAt)),
	}
	for _, event := range events {
		metrics[event] = okay.Count(1)
	}

	// Create the event
	event := &okay.Event{
		Name:      category,
		Labels:    labels,
		Timestamp: startedAt, // Timestamp as start of event
		Metrics:   metrics,
		UniqueKey: []string{"event_id"},
		Properties: map[string]string{
			"event_id": uuid.NewString(),
		},
	}

	// Set to store
	store := getStore(ctx)
	if store != nil {
		store.events = append(store.events, event)
	}

	return event
}

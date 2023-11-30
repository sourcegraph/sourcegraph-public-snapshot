package telemetrytest

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

// MockTelemetryEventsExportQueueStore only implements QueueForExport.
type MockTelemetryEventsExportQueueStore struct {
	database.TelemetryEventsExportQueueStore
	events []*telemetrygatewayv1.Event
}

func NewMockEventsExportQueueStore() *MockTelemetryEventsExportQueueStore {
	return &MockTelemetryEventsExportQueueStore{}
}

func (f *MockTelemetryEventsExportQueueStore) QueueForExport(_ context.Context, events []*telemetrygatewayv1.Event) error {
	f.events = append(f.events, events...)
	return nil
}

type QueuedEvents []*telemetrygatewayv1.Event

// Summary returns a set of strings with format "${feature} - ${action}"
// corresponding to the queued events.
func (q QueuedEvents) Summary() []string {
	var events []string
	for _, e := range q {
		events = append(events, fmt.Sprintf("%s - %s", e.Feature, e.Action))
	}
	return events
}

// GetMockQueuedEvents retrieves the queued events by the mock.
func (f *MockTelemetryEventsExportQueueStore) GetMockQueuedEvents() QueuedEvents {
	return f.events
}

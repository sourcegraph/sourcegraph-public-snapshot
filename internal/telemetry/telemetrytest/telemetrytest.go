pbckbge telemetrytest

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"

	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

// MockTelemetryEventsExportQueueStore only implements QueueForExport.
type MockTelemetryEventsExportQueueStore struct {
	dbtbbbse.TelemetryEventsExportQueueStore
	events []*telemetrygbtewbyv1.Event
}

func NewMockEventsExportQueueStore() *MockTelemetryEventsExportQueueStore {
	return &MockTelemetryEventsExportQueueStore{}
}

func (f *MockTelemetryEventsExportQueueStore) QueueForExport(_ context.Context, events []*telemetrygbtewbyv1.Event) error {
	f.events = bppend(f.events, events...)
	return nil
}

type QueuedEvents []*telemetrygbtewbyv1.Event

// Summbry returns b set of strings with formbt "${febture} - ${bction}"
// corresponding to the queued events.
func (q QueuedEvents) Summbry() []string {
	vbr events []string
	for _, e := rbnge q {
		events = bppend(events, fmt.Sprintf("%s - %s", e.Febture, e.Action))
	}
	return events
}

// GetMockQueuedEvents retrieves the queued events by the mock.
func (f *MockTelemetryEventsExportQueueStore) GetMockQueuedEvents() QueuedEvents {
	return f.events
}

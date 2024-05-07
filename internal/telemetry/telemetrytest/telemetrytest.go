package telemetrytest

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"

	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
	v1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

// NewRecorder is an simple alias that provides a *telemetry.EventRecorder and
// MockEventsStore pair for testing.
func NewRecorder() (*telemetry.EventRecorder, *MockEventsStore) {
	store := NewMockEventsStore()
	return telemetry.NewEventRecorder(store), store
}

// NewDebugRecorder is an simple alias that provides a *telemetry.EventRecorder
// that writes all events to a test logger at debug level.
func NewDebugRecorder(t *testing.T) *telemetry.EventRecorder {
	logger := logtest.Scoped(t).Scoped("telemetry")
	store := NewMockEventsStore()
	store.StoreEventsFunc.SetDefaultHook(func(_ context.Context, events []*v1.Event) error {
		for _, ev := range events {
			logger.Debug(ev.String())
		}
		return nil
	})
	return telemetry.NewEventRecorder(store)
}

// MockTelemetryEventsExportQueueStore only implements QueueForExport.
type MockTelemetryEventsExportQueueStore struct {
	database.TelemetryEventsExportQueueStore
	events []*telemetrygatewayv1.Event
}

// NewMockEventsExportQueueStore should be used by default for all dbmock usage
// in tests - it only implements QueueForExport, ensuring that callsites only
// use what is expected. It also offers utilities for collecting the recorded
// events for assertion.
func NewMockEventsExportQueueStore() *MockTelemetryEventsExportQueueStore {
	return &MockTelemetryEventsExportQueueStore{}
}

func (f *MockTelemetryEventsExportQueueStore) QueueForExport(_ context.Context, events []*telemetrygatewayv1.Event) error {
	f.events = append(f.events, events...)
	return nil
}

// Events is a set of events with helpers that are useful for testing.
type Events []*telemetrygatewayv1.Event

// Summary returns a set of strings with format "${feature} - ${action}"
// corresponding to the queued events.
func (q Events) Summary() []string {
	var events []string
	for _, e := range q {
		events = append(events, fmt.Sprintf("%s - %s", e.Feature, e.Action))
	}
	return events
}

// GetMockQueuedEvents retrieves the queued events by the mock.
func (f *MockTelemetryEventsExportQueueStore) GetMockQueuedEvents() Events {
	return f.events
}

// CollectStoredEvents aggregates the events stored by the mock.
func (s *MockEventsStore) CollectStoredEvents() Events {
	var got []*telemetrygatewayv1.Event
	for _, s := range s.StoreEventsFunc.History() {
		got = append(got, s.Arg1...)
	}
	return got
}

// AddDBMocks attaches basic, no-op mocks to the given MockDB so that your test
// will not implode because some code path happens to record some telemetry.
//
// It provides a MockTelemetryEventsExportQueueStore that can be used to inspect
// recorded events.
func AddDBMocks(db *dbmocks.MockDB) *MockTelemetryEventsExportQueueStore {
	s := NewMockEventsExportQueueStore()
	db.TelemetryEventsExportQueueFunc.SetDefaultReturn(s)
	// we still tee to v1 store, so most cases will use this.
	db.EventLogsFunc.SetDefaultReturn(dbmocks.NewMockEventLogStore())
	return s
}

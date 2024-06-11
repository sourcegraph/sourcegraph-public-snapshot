package telemetrystore

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrystore/teestore"
)

func NewDefaultStore(exportQueue database.TelemetryEventsExportQueueStore, eventLogs database.EventLogStore) telemetry.EventsStore {
	return teestore.NewStore(exportQueue, eventLogs)
}

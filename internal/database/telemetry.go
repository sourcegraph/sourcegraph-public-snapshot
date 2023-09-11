package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

// TelemetryStore is the database for "Telemetry V2", or "event-logging everywhere".
type TelemetryStore interface {
	basestore.ShareableStore

	// RecordEvents persists a set of raw events for the telemetry gateway.
	RecordEvents(ctx context.Context, events []telemetrygatewayv1.Event) error
}

type telemetryStore struct {
	*basestore.Store
}

// TelemetryStoreWith instantiates and returns a new TelemetryStore using the other store handle.
func TelemetryStoreWith(other basestore.ShareableStore) TelemetryStore {
	return &telemetryStore{Store: basestore.NewWithHandle(other.Handle())}
}

// TODO: Currently no-ops.
func (t *telemetryStore) RecordEvents(ctx context.Context, events []telemetrygatewayv1.Event) error {
	return nil
}

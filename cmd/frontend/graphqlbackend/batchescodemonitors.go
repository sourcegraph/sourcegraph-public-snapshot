package graphqlbackend

import "context"

// Problem:
// Those two methods MUST be exposed on the
// BatchChangeResolver and MonitorResolver interfaces.
// Otherwise, graphql-go doesn't pass the type check validation on startup
// and it will panic.
// Also, these two methods would need to be called when BatchChangeResolver.CodeMonitor()
// is called, but that is currently not possible. This 3rd package helps combine the code from
// the two packages into a new set of logic without causing import cycles, but it
// does now allow "injecting" code into those packages.
type BatchesCodeMonitorsResolver interface {
	BatchChangeCodeMonitor(ctx context.Context, monitorID int64) (MonitorResolver, error)
	CodeMonitorBatchChange(ctx context.Context, batchChangeID int64) (BatchChangeResolver, error)
}

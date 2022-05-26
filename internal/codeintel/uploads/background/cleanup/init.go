package cleanup

import (
	"context"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewJanitor(dbStore DBStore, lsifStore LSIFStore, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, newJanitor(
		dbStore,
		lsifStore,
		metrics,
	))
}

func newJanitor(dbStore DBStore, lsifStore LSIFStore, metrics *metrics) *janitor {
	return &janitor{
		dbStore:   dbStore,
		lsifStore: lsifStore,
		metrics:   metrics,
		clock:     glock.NewRealClock(),
	}
}

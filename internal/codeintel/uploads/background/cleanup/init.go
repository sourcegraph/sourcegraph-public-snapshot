package cleanup

import (
	"context"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewJanitor(dbStore DBStore, lsifStore LSIFStore, uploadSvc UploadService, metrics *metrics) goroutine.BackgroundRoutine {
	janitor := newJanitor(dbStore, lsifStore, uploadSvc, metrics)

	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, janitor)
}

func newJanitor(dbStore DBStore, lsifStore LSIFStore, uploadSvc UploadService, metrics *metrics) *janitor {
	return &janitor{
		dbStore:   dbStore,
		lsifStore: lsifStore,
		uploadSvc: uploadSvc,
		metrics:   metrics,
		clock:     glock.NewRealClock(),
	}
}

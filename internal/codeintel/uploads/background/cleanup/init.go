package cleanup

import (
	"context"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewJanitor(dbStore DBStore, uploadSvc UploadService, indexSvc AutoIndexingService, logger log.Logger, metrics *metrics) goroutine.BackgroundRoutine {
	janitor := newJanitor(dbStore, uploadSvc, indexSvc, logger, metrics)

	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, janitor)
}

func newJanitor(dbStore DBStore, uploadSvc UploadService, indexSvc AutoIndexingService, logger log.Logger, metrics *metrics) *janitor {
	return &janitor{
		dbStore:   dbStore,
		uploadSvc: uploadSvc,
		indexSvc:  indexSvc,
		metrics:   metrics,
		clock:     glock.NewRealClock(),
		logger:    logger,
	}
}

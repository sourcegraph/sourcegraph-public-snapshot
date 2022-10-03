package cleanup

import (
	"context"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewJanitor(db database.DB, uploadSvc UploadService, indexSvc AutoIndexingService, logger log.Logger, metrics *metrics) goroutine.BackgroundRoutine {
	janitor := newJanitor(db, uploadSvc, indexSvc, logger, metrics)
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, janitor)
}

func newJanitor(db database.DB, uploadSvc UploadService, indexSvc AutoIndexingService, logger log.Logger, metrics *metrics) *janitor {
	return &janitor{
		gsc:       gitserver.NewClient(db),
		uploadSvc: uploadSvc,
		indexSvc:  indexSvc,
		metrics:   metrics,
		clock:     glock.NewRealClock(),
		logger:    logger,
	}
}

func NewResetters(uploadSvc UploadService, logger log.Logger, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	metrics := newMetrics2(observationContext)

	return []goroutine.BackgroundRoutine{
		NewUploadResetter(logger.Scoped("janitor.UploadResetter", ""), uploadSvc.WorkerutilStore(), ConfigInst.Interval, metrics),
	}
}

package commitgraph

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewUpdater(uploadSvc UploadService, logger log.Logger) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &updater{
		uploadSvc: uploadSvc,
		logger:    logger,
	})
}

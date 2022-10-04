package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type referenceCountUpdater struct {
	uploadSvc UploadServiceForExpiration
	batchSize int
}

var (
	_ goroutine.Handler      = &referenceCountUpdater{}
	_ goroutine.ErrorHandler = &referenceCountUpdater{}
)

func (s *Service) NewReferenceCountUpdater(interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &referenceCountUpdater{
		uploadSvc: s,
		batchSize: batchSize,
	})
}

func (u *referenceCountUpdater) Handle(ctx context.Context) error {
	return u.uploadSvc.BackfillReferenceCountBatch(ctx, u.batchSize)
}

func (u *referenceCountUpdater) HandleError(err error) {
}

package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type committedAtBackfiller struct {
	uploadSvc *Service
	batchSize int
}

var (
	_ goroutine.Handler      = &committedAtBackfiller{}
	_ goroutine.ErrorHandler = &committedAtBackfiller{}
)

func (s *Service) NewCommittedAtBackfiller(interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &committedAtBackfiller{
		uploadSvc: s,
		batchSize: batchSize,
	})
}

func (u *committedAtBackfiller) Handle(ctx context.Context) error {
	return u.uploadSvc.BackfillCommittedAtBatch(ctx, u.batchSize)
}

func (u *committedAtBackfiller) HandleError(err error) {
}

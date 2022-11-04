package backfill

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type committedAtBackfiller struct {
	uploadSvc UploadService
	batchSize int
}

var (
	_ goroutine.Handler      = &committedAtBackfiller{}
	_ goroutine.ErrorHandler = &committedAtBackfiller{}
)

func (u *committedAtBackfiller) Handle(ctx context.Context) error {
	return u.uploadSvc.BackfillCommittedAtBatch(ctx, u.batchSize)
}

func (u *committedAtBackfiller) HandleError(err error) {
}

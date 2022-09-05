package expiration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type referenceCountUpdater struct {
	uploadSvc UploadService
	batchSize int
}

var (
	_ goroutine.Handler      = &referenceCountUpdater{}
	_ goroutine.ErrorHandler = &referenceCountUpdater{}
)

func (u *referenceCountUpdater) Handle(ctx context.Context) error {
	return u.uploadSvc.BackfillReferenceCountBatch(ctx, u.batchSize)
}

func (u *referenceCountUpdater) HandleError(err error) {
}

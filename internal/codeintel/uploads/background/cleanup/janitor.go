package cleanup

import (
	"context"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type janitor struct {
	dbStore   DBStore
	lsifStore LSIFStore
	metrics   *metrics
	clock     glock.Clock
}

var (
	_ goroutine.Handler      = &janitor{}
	_ goroutine.ErrorHandler = &janitor{}
)

func (j *janitor) Handle(ctx context.Context) error {
	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375

	// Reconciliation and denormalization
	j.HandleDeletedRepository(ctx)
	j.HandleUnknownCommit(ctx)

	// Expiration
	j.HandleAbandonedUpload(ctx)
	j.HandleExpiredUploadDeleter(ctx)
	j.HandleHardDeleter(ctx)
	j.HandleAuditLog(ctx)
	return nil
}

func (r *janitor) HandleError(err error) {
}

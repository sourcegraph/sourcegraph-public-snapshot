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
	if err := j.HandleDeletedRepository(ctx); err != nil {
		return err
	}
	if err := j.HandleUnknownCommit(ctx); err != nil {
		return err
	}

	// Expiration
	if err := j.HandleAbandonedUpload(ctx); err != nil {
		return err
	}
	if err := j.HandleExpiredUploadDeleter(ctx); err != nil {
		return err
	}
	if err := j.HandleHardDeleter(ctx); err != nil {
		return err
	}
	if err := j.HandleAuditLog(ctx); err != nil {
		return err
	}

	return nil
}

func (r *janitor) HandleError(err error) {
}

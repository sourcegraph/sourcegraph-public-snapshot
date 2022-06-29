package cleanup

import (
	"context"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type janitor struct {
	logger    log.Logger
	dbStore   DBStore
	lsifStore LSIFStore
	uploadSvc UploadService
	indexSvc  AutoIndexingService
	metrics   *metrics
	clock     glock.Clock
}

var (
	_ goroutine.Handler      = &janitor{}
	_ goroutine.ErrorHandler = &janitor{}
)

func (j *janitor) Handle(ctx context.Context) (errs error) {
	// Reconciliation and denormalization
	if err := j.HandleDeletedRepository(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := j.HandleUnknownCommit(ctx); err != nil {
		errs = errors.Append(errs, err)
	}

	// Expiration
	if err := j.HandleAbandonedUpload(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := j.HandleExpiredUploadDeleter(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := j.HandleHardDeleter(ctx); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := j.HandleAuditLog(ctx); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}

func (r *janitor) HandleError(err error) {
}

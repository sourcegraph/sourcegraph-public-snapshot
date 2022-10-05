package uploads

import (
	"context"
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type janitor struct {
	gsc       GitserverClient
	uploadSvc UploadServiceForCleanup
	indexSvc  AutoIndexingService
	metrics   *janitorMetrics
	logger    log.Logger
	clock     glock.Clock

	uploadTimeout                  time.Duration
	auditLogMaxAge                 time.Duration
	minimumTimeSinceLastCheck      time.Duration
	commitResolverBatchSize        int
	commitResolverMaximumCommitLag time.Duration
}

var (
	_ goroutine.Handler      = &janitor{}
	_ goroutine.ErrorHandler = &janitor{}
)

func (s *Service) NewJanitor(
	interval time.Duration,
	uploadTimeout time.Duration,
	auditLogMaxAge time.Duration,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverBatchSize int,
	commitResolverMaximumCommitLag time.Duration,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &janitor{
		gsc:       s.gitserverClient,
		uploadSvc: s,
		indexSvc:  s.autoIndexingSvc,
		metrics:   s.janitorMetrics,
		clock:     glock.NewRealClock(),
		logger:    s.logger,

		uploadTimeout:                  uploadTimeout,
		auditLogMaxAge:                 auditLogMaxAge,
		minimumTimeSinceLastCheck:      minimumTimeSinceLastCheck,
		commitResolverBatchSize:        commitResolverBatchSize,
		commitResolverMaximumCommitLag: commitResolverMaximumCommitLag,
	})
}

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

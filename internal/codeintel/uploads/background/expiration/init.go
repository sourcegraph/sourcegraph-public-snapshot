package expiration

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewExpirer(uploadSvc UploadService, policySvc PolicyService, policyMatcher PolicyMatcher, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.ExpirerInterval, &expirer{
		uploadSvc:     uploadSvc,
		policySvc:     policySvc,
		policyMatcher: policyMatcher,
		metrics:       metrics,
		logger:        log.Scoped("Expirer", ""),
	})
}

func NewReferenceCountUpdater(uploadSvc UploadService) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.ReferenceCountUpdaterInterval, &referenceCountUpdater{
		uploadSvc: uploadSvc,
		batchSize: ConfigInst.ReferenceCountUpdaterBatchSize,
	})
}

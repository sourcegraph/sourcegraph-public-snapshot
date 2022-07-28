package expiration

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewExpirer(uploadSvc UploadService, policySvc PolicyService, policyMatcher PolicyMatcher, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &expirer{
		uploadSvc:     uploadSvc,
		policySvc:     policySvc,
		policyMatcher: policyMatcher,
		metrics:       metrics,
		logger:        log.Scoped("Expirer", ""),
	})
}

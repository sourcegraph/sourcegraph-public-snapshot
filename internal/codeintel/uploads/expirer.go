package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type expirer struct {
	uploadSvc     UploadServiceForExpiration
	policySvc     PolicyService
	metrics       *expirationMetrics
	policyMatcher PolicyMatcher
	logger        log.Logger

	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	uploadProcessDelay     time.Duration
	uploadBatchSize        int
	commitBatchSize        int
	policyBatchSize        int
}

var (
	_ goroutine.Handler      = &expirer{}
	_ goroutine.ErrorHandler = &expirer{}
)

func (s *Service) NewExpirer(
	interval time.Duration,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	uploadProcessDelay time.Duration,
	uploadBatchSize int,
	commitBatchSize int,
	policyBatchSize int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &expirer{
		uploadSvc:     s,
		policySvc:     s.policySvc,
		policyMatcher: s.policyMatcher,
		metrics:       s.expirationMetrics,
		logger:        log.Scoped("Expirer", ""),

		repositoryProcessDelay: repositoryProcessDelay,
		repositoryBatchSize:    repositoryBatchSize,
		uploadProcessDelay:     uploadProcessDelay,
		uploadBatchSize:        uploadBatchSize,
		commitBatchSize:        commitBatchSize,
		policyBatchSize:        policyBatchSize,
	})
}

func (r *expirer) Handle(ctx context.Context) error {
	if err := r.HandleUploadExpirer(ctx); err != nil {
		return err
	}

	return nil
}

func (r *expirer) HandleError(err error) {
}

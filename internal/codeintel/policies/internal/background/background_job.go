package background

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type BackgroundJob interface {
	NewRepositoryMatcher(interval time.Duration, configurationPolicyMembershipBatchSize int) goroutine.BackgroundRoutine
	SetPolicyService(svc PolicyService)
}

type backgroundJob struct {
	policySvc      PolicyService
	matcherMetrics *matcherMetrics
}

func New(observationContext *observation.Context) BackgroundJob {
	return &backgroundJob{
		matcherMetrics: newMetrics(observationContext),
	}
}

func (b *backgroundJob) SetPolicyService(svc PolicyService) {
	b.policySvc = svc
}

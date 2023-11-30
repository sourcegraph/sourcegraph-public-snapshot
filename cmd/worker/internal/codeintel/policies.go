package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type policiesRepositoryMatcherJob struct{}

func NewPoliciesRepositoryMatcherJob() job.Job {
	return &policiesRepositoryMatcherJob{}
}

func (j *policiesRepositoryMatcherJob) Description() string {
	return "code-intel policies repository matcher"
}

func (j *policiesRepositoryMatcherJob) Config() []env.Config {
	return []env.Config{
		policies.RepositoryMatcherConfigInst,
	}
}

func (j *policiesRepositoryMatcherJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	// TODO(nsc): https://github.com/sourcegraph/sourcegraph/pull/42765
	return policies.NewRepositoryMatcherRoutines(observationCtx, services.PoliciesService), nil
}

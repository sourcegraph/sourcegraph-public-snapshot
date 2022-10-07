package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type policiesRepositoryMatcherJob struct {
	observationContext *observation.Context
}

func NewPoliciesRepositoryMatcherJob(observationContext *observation.Context) job.Job {
	return &policiesRepositoryMatcherJob{observationContext: &observation.Context{
		Logger:       log.NoOp(),
		Tracer:       observationContext.Tracer,
		Registerer:   observationContext.Registerer,
		HoneyDataset: observationContext.HoneyDataset,
	}}
}

func (j *policiesRepositoryMatcherJob) Description() string {
	return "code-intel policies repository matcher"
}

func (j *policiesRepositoryMatcherJob) Config() []env.Config {
	return []env.Config{
		policies.PolicyMatcherConfigInst,
	}
}

func (j *policiesRepositoryMatcherJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	// TODO(nsc): https://github.com/sourcegraph/sourcegraph/pull/42765
	return policies.PolicyMatcherJobs(services.PoliciesService, observation.ScopedContext("codeintel", "policies", "repoMatcherJob", j.observationContext)), nil
}

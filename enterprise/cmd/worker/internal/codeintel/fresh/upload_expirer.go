package codeintel

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/expiration"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type uploadExpirerJob struct{}

func NewUploadExpirerJob() job.Job {
	return &uploadExpirerJob{}
}

func (j *uploadExpirerJob) Description() string {
	return ""
}

func (j *uploadExpirerJob) Config() []env.Config {
	return []env.Config{
		expiration.ConfigInst,
	}
}

func (j *uploadExpirerJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "codeintel job routines"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	metrics := expiration.NewMetrics(observationContext)

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	policyMatcher := policies.NewMatcher(gitserverClient, policies.RetentionExtractor, true, false)

	return []goroutine.BackgroundRoutine{
		expiration.NewExpirer(expiration.DBStoreShim{Store: dbStore}, policyMatcher, metrics),
	}, nil
}

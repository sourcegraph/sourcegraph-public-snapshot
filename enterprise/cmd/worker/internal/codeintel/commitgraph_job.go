package codeintel

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type commitGraphJob struct{}

func NewCommitGraphJob() job.Job {
	return &commitGraphJob{}
}

func (j *commitGraphJob) Description() string {
	return ""
}

func (j *commitGraphJob) Config() []env.Config {
	return []env.Config{commitGraphConfigInst}
}

func (j *commitGraphJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "commit graph job routines"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	locker := locker.NewWithDB(db, "codeintel")

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		commitgraph.NewUpdater(
			dbStore,
			locker,
			gitserverClient,
			commitGraphConfigInst.MaxAgeForNonStaleBranches,
			commitGraphConfigInst.MaxAgeForNonStaleTags,
			commitGraphConfigInst.CommitGraphUpdateTaskInterval,
			observationContext,
		),
	}

	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_queued_duration_seconds_total",
		Help: "The maximum amount of time a repository has had a stale commit graph.",
	}, func() float64 {
		age, err := dbStore.MaxStaleAge(context.Background())
		if err != nil {
			log15.Error("Failed to determine stale commit graph age", "error", err)
			return 0
		}

		return float64(age) / float64(time.Second)
	}))

	return routines, nil
}

package codeintel

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type commitGraphJob struct{}

func NewCommitGraphJob() shared.Job {
	return &commitGraphJob{}
}

func (j *commitGraphJob) Config() []env.Config {
	return []env.Config{commitGraphConfigInst}
}

func (j *commitGraphJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := shared.InitDatabase()
	if err != nil {
		return nil, err
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	locker := locker.NewWithDB(db, "codeintel")

	gitserverClient, err := InitGitserverClient()
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		commitgraph.NewUpdater(dbStore, locker, gitserverClient, commitGraphConfigInst.CommitGraphUpdateTaskInterval, observationContext),
	}

	return routines, nil
}

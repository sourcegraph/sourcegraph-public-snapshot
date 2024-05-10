package syntactic_indexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type syntacticIndexingSchedulerJob struct{}

var _ job.Job = &syntacticIndexingSchedulerJob{}

var config *SchedulerConfig = &SchedulerConfig{}

func NewSyntacticindexingSchedulerJob() job.Job {
	return &syntacticIndexingSchedulerJob{}
}

func (j *syntacticIndexingSchedulerJob) Description() string {
	return ""
}

func (j *syntacticIndexingSchedulerJob) Config() []env.Config {
	return []env.Config{
		config,
	}
}

func (j *syntacticIndexingSchedulerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	scheduler, err := NewSyntacticJobScheduler(observationCtx)

	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		newSchedulerJob(
			observationCtx,
			scheduler,
		),
	}, nil

}

func newSchedulerJob(
	observationCtx *observation.Context,
	scheduler SyntacticJobScheduler,
) goroutine.BackgroundRoutine {

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return scheduler.Schedule(observationCtx, ctx, time.Now())
		}),
		goroutine.WithName("codeintel.autoindexing-background-scheduler"),
		goroutine.WithDescription("schedule autoindexing jobs in the background using defined or inferred configurations"),
		goroutine.WithInterval(time.Second*5),
		goroutine.WithOperation(observationCtx.Operation(observation.Op{
			Name:              "codeintel.indexing.HandleIndexSchedule",
			MetricLabelValues: []string{"HandleIndexSchedule"},
			// Metrics:           redMetrics,
			// ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			// 	if errors.As(err, &inference.LimitError{}) {
			// 		return observation.EmitForNone
			// 	}
			// 	return observation.EmitForDefault
			// },
		})),
	)
}

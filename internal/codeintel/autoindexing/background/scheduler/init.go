package scheduler

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewScheduler(
	autoindexingSvc *autoindexing.Service,
	policySvc PolicyService,
	uploadSvc UploadService,
	policyMatcher PolicyMatcher,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_index_scheduler",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	handleIndexScheduler := observationContext.Operation(observation.Op{
		Name:              "codeintel.indexing.HandleIndexSchedule",
		MetricLabelValues: []string{"HandleIndexSchedule"},
		Metrics:           m,
		ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			if errors.As(err, &inference.LimitError{}) {
				return observation.EmitForDefault.Without(observation.EmitForMetrics)
			}
			return observation.EmitForDefault
		},
	})

	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), ConfigInst.Interval, &scheduler{
		autoindexingSvc: autoindexingSvc,
		policySvc:       policySvc,
		uploadSvc:       uploadSvc,
		policyMatcher:   policyMatcher,
	}, handleIndexScheduler)
}

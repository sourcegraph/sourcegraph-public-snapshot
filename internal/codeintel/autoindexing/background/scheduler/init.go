package scheduler

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewSchedulers(backgroundJobs AutoIndexingServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewScheduler(
			ConfigInst.SchedulerInterval,
			ConfigInst.RepositoryProcessDelay,
			ConfigInst.RepositoryBatchSize,
			ConfigInst.PolicyBatchSize,
		),

		backgroundJobs.NewOnDemandScheduler(
			ConfigInst.OnDemandSchedulerInterval,
			ConfigInst.OnDemandBatchsize,
		),
	}
}

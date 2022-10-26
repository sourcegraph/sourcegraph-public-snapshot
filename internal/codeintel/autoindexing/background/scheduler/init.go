package scheduler

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewSchedulers(backgroundJobs AutoIndexingServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewRepoIndexingScheduler(
			ConfigInst.SchedulerInterval,
			ConfigInst.RepositoryProcessDelay,
			ConfigInst.RepositoryBatchSize,
			ConfigInst.PolicyBatchSize,
			ConfigInst.InferenceConcurrency,
		),

		backgroundJobs.NewOnDemandScheduler(
			ConfigInst.OnDemandSchedulerInterval,
			ConfigInst.OnDemandBatchsize,
		),
	}
}

package scheduler

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewSchedulers(autoIndexingSvc AutoIndexingService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		autoIndexingSvc.NewScheduler(
			ConfigInst.SchedulerInterval,
			ConfigInst.RepositoryProcessDelay,
			ConfigInst.RepositoryBatchSize,
			ConfigInst.PolicyBatchSize,
		),

		autoIndexingSvc.NewOnDemandScheduler(
			ConfigInst.OnDemandSchedulerInterval,
			ConfigInst.OnDemandBatchsize,
		),
	}
}

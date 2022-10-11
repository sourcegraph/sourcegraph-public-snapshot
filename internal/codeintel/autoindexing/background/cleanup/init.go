package cleanup

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewJanitor(autoIndeingSvc AutoIndexingService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		autoIndeingSvc.NewJanitor(
			ConfigInst.Interval,
			ConfigInst.MinimumTimeSinceLastCheck,
			ConfigInst.CommitResolverBatchSize,
			ConfigInst.CommitResolverMaximumCommitLag,
		),
	}
}

func NewResetters(backgroundJobs AutoIndexingServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewIndexResetter(ConfigInst.Interval),
		backgroundJobs.NewDependencyIndexResetter(ConfigInst.Interval),
	}
}

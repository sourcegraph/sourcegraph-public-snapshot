package cleanup

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewJanitor(backgroundJobs AutoIndexingServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewJanitor(
			ConfigInst.Interval,
			ConfigInst.MinimumTimeSinceLastCheck,
			ConfigInst.CommitResolverBatchSize,
			ConfigInst.CommitResolverMaximumCommitLag,
			ConfigInst.FailedIndexBatchSize,
			ConfigInst.FailedIndexMaxAge,
		),
	}
}

func NewResetters(backgroundJobs AutoIndexingServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewIndexResetter(ConfigInst.Interval),
		backgroundJobs.NewDependencyIndexResetter(ConfigInst.Interval),
	}
}

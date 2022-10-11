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

func NewResetters(autoIndexingSvc AutoIndexingService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		autoIndexingSvc.NewIndexResetter(ConfigInst.Interval),
		autoIndexingSvc.NewDependencyIndexResetter(ConfigInst.Interval),
	}
}

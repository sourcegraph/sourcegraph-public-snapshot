package dependencies

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewSchedulers(autoIndexingSvc AutoIndexingService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		autoIndexingSvc.NewDependencySyncScheduler(ConfigInst.DependencyIndexerSchedulerPollInterval),
		autoIndexingSvc.NewDependencyIndexingScheduler(ConfigInst.DependencyIndexerSchedulerPollInterval, ConfigInst.DependencyIndexerSchedulerConcurrency),
	}
}

package dependencies

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewSchedulers(backgroundJobs AutoIndexingServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewDependencySyncScheduler(ConfigInst.DependencyIndexerSchedulerPollInterval),
		backgroundJobs.NewDependencyIndexingScheduler(ConfigInst.DependencyIndexerSchedulerPollInterval, ConfigInst.DependencyIndexerSchedulerConcurrency),
	}
}

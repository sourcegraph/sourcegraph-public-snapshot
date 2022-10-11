package cleanup

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewResetters(backgroundJobs AutoIndexingServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewIndexResetter(ConfigInst.Interval),
		backgroundJobs.NewDependencyIndexResetter(ConfigInst.Interval),
	}
}

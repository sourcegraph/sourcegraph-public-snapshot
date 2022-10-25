package cratesyncer

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewCrateSyncer(depsSvcBackgroundJobs DependenciesServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		depsSvcBackgroundJobs.NewCrateSyncer(),
	}
}

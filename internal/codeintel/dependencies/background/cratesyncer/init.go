package cratesyncer

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewSyncers(depsSvc DependenciesService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		depsSvc.NewCrateSyncer(),
	}
}

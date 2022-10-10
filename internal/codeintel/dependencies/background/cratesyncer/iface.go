package cratesyncer

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

type DependenciesService interface {
	NewCrateSyncer() goroutine.BackgroundRoutine
}

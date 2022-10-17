package cratesyncer

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

type DependenciesServiceBackgroundJobs interface {
	NewCrateSyncer() goroutine.BackgroundRoutine
}

package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	deleteDependencyReposByID    *observation.Operation
	listDependencyRepos          *observation.Operation
	lockfileDependencies         *observation.Operation
	lockfileDependents           *observation.Operation
	preciseDependencies          *observation.Operation
	preciseDependents            *observation.Operation
	selectRepoRevisionsToResolve *observation.Operation
	updateResolvedRevisions      *observation.Operation
	upsertDependencyRepos        *observation.Operation
	upsertLockfileGraph          *observation.Operation
	listLockfileIndexes          *observation.Operation
	getLockfileIndex             *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_dependencies_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dependencies.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &operations{
		deleteDependencyReposByID:    op("DeleteDependencyReposByID"),
		listDependencyRepos:          op("ListDependencyRepos"),
		lockfileDependencies:         op("LockfileDependencies"),
		lockfileDependents:           op("LockfileDependents"),
		preciseDependencies:          op("PreciseDependencies"),
		preciseDependents:            op("PreciseDependents"),
		selectRepoRevisionsToResolve: op("SelectRepoRevisionsToResolve"),
		updateResolvedRevisions:      op("UpdateResolvedRevisions"),
		upsertDependencyRepos:        op("InsertDependencyRepos"),
		upsertLockfileGraph:          op("UpsertLockfileGraph"),
		listLockfileIndexes:          op("ListLockfileIndexes"),
		getLockfileIndex:             op("GetLockfileIndex"),
	}
}

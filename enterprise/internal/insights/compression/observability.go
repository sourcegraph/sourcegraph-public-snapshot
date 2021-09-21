package compression

import (
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	Worker     *observation.Operation
	GetCommits *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	getCommits := observationContext.Operation(observation.Op{
		Name: "CommitIndexer.GetCommits",
		Metrics: metrics.NewOperationMetrics(
			observationContext.Registerer,
			"src_insights_commit_indexer_fetch_duration",
			metrics.WithCountHelp("Time for the commit indexer to fetch commits from gitserver."),
		),
	})

	worker := observationContext.Operation(observation.Op{
		Name: "CommitIndexer.Run",
		Metrics: metrics.NewOperationMetrics(
			observationContext.Registerer,
			"commit_indexer",
			metrics.WithCountHelp("Total number of commit indexer executions"),
		),
	})

	return &operations{
		Worker:     worker,
		GetCommits: getCommits,
	}
}

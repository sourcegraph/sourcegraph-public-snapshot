package compression

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	worker       *observation.Operation
	getCommits   *observation.Operation
	countCommits *prometheus.CounterVec
}

func newOperations(observationCtx *observation.Context) *operations {
	worker := observationCtx.Operation(observation.Op{
		Name: "CommitIndexer.Run",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"insights_commit_indexer",
			metrics.WithCountHelp("Total number of commit indexer executions"),
		),
	})

	getCommits := observationCtx.Operation(observation.Op{
		Name: "CommitIndexer.GetCommits",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"insights_commit_indexer_fetch",
			metrics.WithCountHelp("Time for the commit indexer to fetch commits from gitserver."),
		),
	})

	countCommits := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "insights_commit_indexer_commits_added",
		Help: "Number of commits added to the commit index",
	}, []string{})

	return &operations{
		worker:       worker,
		getCommits:   getCommits,
		countCommits: countCommits,
	}
}

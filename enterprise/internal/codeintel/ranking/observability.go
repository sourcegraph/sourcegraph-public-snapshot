package ranking

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getRepoRank       *observation.Operation
	getDocumentRanks  *observation.Operation
	indexRepositories *observation.Operation
	indexRepository   *observation.Operation

	numCSVFilesProcessed   prometheus.Counter
	numCSVBytesRead        prometheus.Counter
	numRepositoriesUpdated prometheus.Counter
	numInputRowsProcessed  prometheus.Counter
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_ranking",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.ranking.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numCSVFilesProcessed := counter(
		"src_codeintel_ranking_csv_files_processed_total",
		"The number of input CSV records read from GCS.",
	)
	numCSVBytesRead := counter(
		"src_codeintel_ranking_csv_files_bytes_read_total",
		"The number of bytes read from GCS.",
	)
	numRepositoriesUpdated := counter(
		"src_codeintel_ranking_repositories_updated_total",
		"The number of updates to document scores of any repository.",
	)
	numInputRowsProcessed := counter(
		"src_codeintel_ranking_input_rows_processed_total",
		"The number of input row records merged into document scores for a single repo.",
	)

	return &operations{
		getRepoRank:       op("GetRepoRank"),
		getDocumentRanks:  op("GetDocumentRanks"),
		indexRepositories: op("IndexRepositories"),
		indexRepository:   op("indexRepository"),

		numCSVFilesProcessed:   numCSVFilesProcessed,
		numCSVBytesRead:        numCSVBytesRead,
		numRepositoriesUpdated: numRepositoriesUpdated,
		numInputRowsProcessed:  numInputRowsProcessed,
	}
}

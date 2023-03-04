package ranking

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	exportRankingGraph *observation.Operation
	mapRankingGraph    *observation.Operation
	reduceRankingGraph *observation.Operation
	getRepoRank        *observation.Operation
	getDocumentRanks   *observation.Operation
	indexRepositories  *observation.Operation
	indexRepository    *observation.Operation

	numUploadsRead                   prometheus.Counter
	numDefinitionsInserted           prometheus.Counter
	numReferencesInserted            prometheus.Counter
	numStaleDefinitionRecordsDeleted prometheus.Counter
	numStaleReferenceRecordsDeleted  prometheus.Counter
	numMetadataRecordsDeleted        prometheus.Counter
	numInputRecordsDeleted           prometheus.Counter
	numRankRecordsDeleted            prometheus.Counter
}

var (
	metricsMap = make(map[string]prometheus.Counter)
	m          = new(metrics.SingletonREDMetrics)
	metricsMu  sync.Mutex
)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_ranking",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.ranking.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	counter := func(name, help string) prometheus.Counter {
		metricsMu.Lock()
		defer metricsMu.Unlock()

		if c, ok := metricsMap[name]; ok {
			return c
		}

		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})
		observationCtx.Registerer.MustRegister(counter)

		metricsMap[name] = counter

		return counter
	}

	numUploadsRead := counter(
		"src_codeintel_ranking_ranking_uploads_read_total",
		"The number of upload records read.",
	)
	numDefinitionsInserted := counter(
		"src_codeintel_ranking_num_definitions_inserted_total",
		"The number of definition records inserted into Postgres.",
	)
	numReferencesInserted := counter(
		"src_codeintel_ranking_num_references_inserted_total",
		"The number of reference records inserted into Postgres.",
	)
	numStaleDefinitionRecordsDeleted := counter(
		"src_codeintel_ranking_num_stale_definition_records_deleted_total",
		"The number of stale definition records removed from Postgres.",
	)
	numStaleReferenceRecordsDeleted := counter(
		"src_codeintel_ranking_num_stale_reference_records_deleted_total",
		"The number of stale reference records removed from Postgres.",
	)
	numMetadataRecordsDeleted := counter(
		"src_codeintel_ranking_num_metadata_records_deleted_total",
		"The number of stale metadata records removed from Postgres.",
	)
	numInputRecordsDeleted := counter(
		"src_codeintel_ranking_num_input_records_deleted_total",
		"The number of stale input records removed from Postgres.",
	)
	numRankRecordsDeleted := counter(
		"src_codeintel_ranking_num_rank_records_deleted_total",
		"The number of stale rank records removed from Postgres.",
	)

	return &operations{
		exportRankingGraph: op("ExportRankingGraph"),
		mapRankingGraph:    op("MapRankingGraph"),
		reduceRankingGraph: op("ReduceRankingGraph"),
		getRepoRank:        op("GetRepoRank"),
		getDocumentRanks:   op("GetDocumentRanks"),
		indexRepositories:  op("IndexRepositories"),
		indexRepository:    op("indexRepository"),

		numUploadsRead:                   numUploadsRead,
		numDefinitionsInserted:           numDefinitionsInserted,
		numReferencesInserted:            numReferencesInserted,
		numStaleDefinitionRecordsDeleted: numStaleDefinitionRecordsDeleted,
		numStaleReferenceRecordsDeleted:  numStaleReferenceRecordsDeleted,
		numMetadataRecordsDeleted:        numMetadataRecordsDeleted,
		numInputRecordsDeleted:           numInputRecordsDeleted,
		numRankRecordsDeleted:            numRankRecordsDeleted,
	}
}

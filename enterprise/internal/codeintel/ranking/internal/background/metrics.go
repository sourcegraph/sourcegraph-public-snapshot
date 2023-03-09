package background

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
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
	metricsMu  sync.Mutex
)

func NewMetrics(observationCtx *observation.Context) *metrics {
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

	return &metrics{
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

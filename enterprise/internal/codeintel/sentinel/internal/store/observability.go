package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	vulnerabilityByID                        *observation.Operation
	getVulnerabilitiesByIDs                  *observation.Operation
	getVulnerabilities                       *observation.Operation
	getVulnerabilityMatchesByRepo            *observation.Operation
	getVulnerabilityMatchesSummaryCounts     *observation.Operation
	getVulnerabilityMatchesCountByRepository *observation.Operation
	getVulnerabilityMatchesSummaryCounts     *observation.Operation
	insertVulnerabilities                    *observation.Operation
	vulnerabilityMatchByID                   *observation.Operation
	getVulnerabilityMatches                  *observation.Operation
	scanMatches                              *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_sentinel_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.sentinel.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		vulnerabilityByID:                        op("VulnerabilityByID"),
		getVulnerabilitiesByIDs:                  op("GetVulnerabilitiesByIDs"),
		getVulnerabilities:                       op("GetVulnerabilities"),
		getVulnerabilityMatchesSummaryCounts:     op("GetVulnerabilityMatchesSummaryCounts"),
		getVulnerabilityMatchesCountByRepository: op("GetVulnerabilityMatchesCountByRepository"),
		getVulnerabilityMatchesByRepo:            op("GetVulnerabilityMatchesByRepo"),
		getVulnerabilityMatchesSummaryCounts:     op("GetVulnerabilityMatchesSummaryCounts"),
		insertVulnerabilities:                    op("InsertVulnerabilities"),
		vulnerabilityMatchByID:                   op("VulnerabilityMatchByID"),
		getVulnerabilityMatches:                  op("GetVulnerabilityMatches"),
		scanMatches:                              op("ScanMatches"),
	}
}

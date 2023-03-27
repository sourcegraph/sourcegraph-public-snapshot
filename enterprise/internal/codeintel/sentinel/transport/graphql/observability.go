package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getVulnerabilities                    *observation.Operation
	vulnerabilityByID                     *observation.Operation
	getMatches                            *observation.Operation
	vulnerabilityMatchByID                *observation.Operation
	vulnerabilityMatchesSummaryCounts     *observation.Operation
	vulnerabilityMatchesCountByRepository *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_sentinel_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.sentinel.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		getVulnerabilities:                    op("Vulnerabilities"),
		vulnerabilityByID:                     op("VulnerabilityByID"),
		getMatches:                            op("Matches"),
		vulnerabilityMatchByID:                op("VulnerabilityMatchByID"),
		vulnerabilityMatchesSummaryCounts:     op("VulnerabilityMatchesSummaryCounts"),
		vulnerabilityMatchesCountByRepository: op("VulnerabilityMatchesCountByRepository"),
	}
}

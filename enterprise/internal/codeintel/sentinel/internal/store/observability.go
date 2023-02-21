package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getVulnerabilities      *observation.Operation
	insertVulnerabilities   *observation.Operation
	getVulnerabilityMatches *observation.Operation
	scanMatches             *observation.Operation
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
		getVulnerabilities:      op("GetVulnerabilities"),
		insertVulnerabilities:   op("InsertVulnerabilities"),
		getVulnerabilityMatches: op("GetVulnerabilityMatches"),
		scanMatches:             op("ScanMatches"),
	}
}

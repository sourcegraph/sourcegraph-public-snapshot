package dbstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	repoName           *observation.Operation
	getJVMDependencies *observation.Operation
	getNPMDependencies *observation.Operation
}

func NewREDMetrics(observationContext *observation.Context) *metrics.REDMetrics {
	return metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_dbstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)
}

func NewOperations(observationContext *observation.Context, metrics *metrics.REDMetrics) *Operations {
	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dbstore.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &Operations{
		repoName:           op("RepoName"),
		getJVMDependencies: op("GetJVMDependencies"),
		getNPMDependencies: op("GetNPMDependencies"),
	}
}

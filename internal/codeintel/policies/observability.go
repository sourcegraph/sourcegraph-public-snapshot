package policies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	commitsMatchingIndexingPolicies  *observation.Operation
	commitsMatchingRetentionPolicies *observation.Operation
	create                           *observation.Operation
	delete                           *observation.Operation
	get                              *observation.Operation
	list                             *observation.Operation
	update                           *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_policies",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.policies.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		commitsMatchingIndexingPolicies:  op("CommitsMatchingIndexingPolicies"),
		commitsMatchingRetentionPolicies: op("CommitsMatchingRetentionPolicies"),
		create:                           op("Create"),
		delete:                           op("Delete"),
		get:                              op("Get"),
		list:                             op("List"),
		update:                           op("Update"),
	}
}

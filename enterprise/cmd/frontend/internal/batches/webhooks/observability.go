package webhooks

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	bitbucketServeHTTP *observation.Operation
}

func NewREDMetrics(observationContext *observation.Context) *metrics.REDMetrics {
	return metrics.NewREDMetrics(
		observationContext.Registerer,
		"batches_webhooks",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)
}

func NewOperations(observationContext *observation.Context, metrics *metrics.REDMetrics) *Operations {
	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("batches.webhooks.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &Operations{
		bitbucketServeHTTP: op("bitbucketServeHTTP"),
	}
}

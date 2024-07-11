package gitserver

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	deleteRepository *observation.Operation
	fetchRepository  *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"gitserver_repositoryservice_client",
		metrics.WithLabels("op", "scope"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("reposervice.client.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				return observation.EmitForAllExceptLogs
			},
		})
	}

	return &operations{
		deleteRepository: op("DeleteRepository"),
		fetchRepository:  op("FetchRepository"),
	}
}

var (
	operationsInst     *operations
	operationsInstOnce sync.Once
)

func getOperations() *operations {
	operationsInstOnce.Do(func() {
		observationCtx := observation.NewContext(log.Scoped("reposervice.client"))
		operationsInst = newOperations(observationCtx)
	})

	return operationsInst
}

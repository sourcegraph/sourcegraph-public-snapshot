package command

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	GitInit           *observation.Operation
	GitFetch          *observation.Operation
	GitCheckout       *observation.Operation
	DockerPull        *observation.Operation
	DockerSave        *observation.Operation
	FirecrackerStart  *observation.Operation
	DockerLoad        *observation.Operation
	SetupRm           *observation.Operation
	FirecrackerStop   *observation.Operation
	FirecrackerRemove *observation.Operation
	DockerRun         *observation.Operation
	IgniteExec        *observation.Operation
}

func MakeOperations(observationContext *observation.Context) *Operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"apiworker_command",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(opName string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         fmt.Sprintf("apiworker.%s", opName),
			MetricLabels: []string{opName},
			Metrics:      metrics,
		})
	}

	return &Operations{
		DockerPull:        op("docker-pull"),
		DockerSave:        op("docker-save"),
		FirecrackerStart:  op("firecracker-start"),
		DockerLoad:        op("docker-load"),
		SetupRm:           op("setup-rm"),
		FirecrackerStop:   op("firecracker-stop"),
		FirecrackerRemove: op("firecracker-remove"),
		DockerRun:         op("docker-run"),
		IgniteExec:        op("ignite-exec"),
		GitInit:           op("git-init"),
		GitFetch:          op("git-fetch"),
		GitCheckout:       op("git-checkout"),
	}
}

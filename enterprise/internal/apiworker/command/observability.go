package command

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	GitInit           *observation.Operation
	GitFetch          *observation.Operation
	GitCheckout       *observation.Operation
	DockerPull        *observation.Operation
	DockerSave        *observation.Operation
	DockerLoad        *observation.Operation
	FirecrackerStart  *observation.Operation
	SetupRm           *observation.Operation
	FirecrackerStop   *observation.Operation
	FirecrackerRemove *observation.Operation
	Exec              *observation.Operation
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
			MetricLabels: []string{strings.Replace(opName, ".", "_", -1)},
			Metrics:      metrics,
		})
	}

	return &Operations{
		GitInit:           op("setup.git.init"),
		GitFetch:          op("setup.git.fetch"),
		GitCheckout:       op("setup.git.checkout"),
		DockerPull:        op("setup.docker.pull"),
		DockerSave:        op("setup.docker.save"),
		DockerLoad:        op("setup.docker.load"),
		SetupRm:           op("setup.rm"),
		FirecrackerStart:  op("setup.firecracker.start"),
		FirecrackerStop:   op("teardown.firecracker.stop"),
		FirecrackerRemove: op("teardown.firecracker.remove"),
		Exec:              op("exec"),
	}
}

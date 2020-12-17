package command

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	SetupGitInit              *observation.Operation
	SetupGitFetch             *observation.Operation
	SetupGitCheckout          *observation.Operation
	SetupDockerPull           *observation.Operation
	SetupDockerSave           *observation.Operation
	SetupDockerLoad           *observation.Operation
	SetupFirecrackerStart     *observation.Operation
	SetupRm                   *observation.Operation
	TeardownFirecrackerStop   *observation.Operation
	TeardownFirecrackerRemove *observation.Operation
	Exec                      *observation.Operation
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
		SetupGitInit:              op("setup.git.init"),
		SetupGitFetch:             op("setup.git.fetch"),
		SetupGitCheckout:          op("setup.git.checkout"),
		SetupDockerPull:           op("setup.docker.pull"),
		SetupDockerSave:           op("setup.docker.save"),
		SetupDockerLoad:           op("setup.docker.load"),
		SetupRm:                   op("setup.rm"),
		SetupFirecrackerStart:     op("setup.firecracker.start"),
		TeardownFirecrackerStop:   op("teardown.firecracker.stop"),
		TeardownFirecrackerRemove: op("teardown.firecracker.remove"),
		Exec:                      op("exec"),
	}
}

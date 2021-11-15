package command

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	SetupGitInit              *observation.Operation
	SetupGitFetch             *observation.Operation
	SetupAddRemote            *observation.Operation
	SetupGitCheckout          *observation.Operation
	SetupFirecrackerStart     *observation.Operation
	SetupStartupScript        *observation.Operation
	TeardownFirecrackerRemove *observation.Operation
	Exec                      *observation.Operation

	RunLockWaitTotal prometheus.Counter
	RunLockHeldTotal prometheus.Counter
}

func NewOperations(observationContext *observation.Context) *Operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"apiworker_command",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(opName string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("apiworker.%s", opName),
			MetricLabelValues: []string{opName},
			Metrics:           metrics,
		})
	}

	runLockWaitTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_executor_run_lock_wait_total",
		Help: "The number of milliseconds spent waiting for the run lock.",
	})
	observationContext.Registerer.MustRegister(runLockWaitTotal)

	runLockHeldTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_executor_run_lock_held_total",
		Help: "The number of milliseconds spent holding the run lock.",
	})
	observationContext.Registerer.MustRegister(runLockHeldTotal)

	return &Operations{
		SetupGitInit:              op("setup.git.init"),
		SetupGitFetch:             op("setup.git.fetch"),
		SetupAddRemote:            op("setup.git.add-remote"),
		SetupGitCheckout:          op("setup.git.checkout"),
		SetupFirecrackerStart:     op("setup.firecracker.start"),
		SetupStartupScript:        op("setup.startup-script"),
		TeardownFirecrackerRemove: op("teardown.firecracker.remove"),
		Exec:                      op("exec"),

		RunLockWaitTotal: runLockWaitTotal,
		RunLockHeldTotal: runLockHeldTotal,
	}
}

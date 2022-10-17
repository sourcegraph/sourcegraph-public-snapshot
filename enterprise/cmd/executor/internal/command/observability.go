package command

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	SetupGitInit                 *observation.Operation
	SetupAddRemote               *observation.Operation
	SetupGitDisableGC            *observation.Operation
	SetupGitFetch                *observation.Operation
	SetupGitSparseCheckoutConfig *observation.Operation
	SetupGitSparseCheckoutSet    *observation.Operation
	SetupGitCheckout             *observation.Operation
	SetupGitSetRemoteUrl         *observation.Operation
	SetupFirecrackerStart        *observation.Operation
	SetupStartupScript           *observation.Operation
	TeardownFirecrackerRemove    *observation.Operation
	Exec                         *observation.Operation

	RunLockWaitTotal prometheus.Counter
	RunLockHeldTotal prometheus.Counter
}

func NewOperations(observationContext *observation.Context) *Operations {
	metrics := metrics.NewREDMetrics(
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
		SetupGitInit:                 op("setup.git.init"),
		SetupAddRemote:               op("setup.git.add-remote"),
		SetupGitDisableGC:            op("setup.git.disable-gc"),
		SetupGitFetch:                op("setup.git.fetch"),
		SetupGitSparseCheckoutConfig: op("setup.git.sparse-checkout-config"),
		SetupGitSparseCheckoutSet:    op("setup.git.sparse-checkout-set"),
		SetupGitCheckout:             op("setup.git.checkout"),
		SetupGitSetRemoteUrl:         op("setup.git.set-remote"),
		SetupFirecrackerStart:        op("setup.firecracker.start"),
		SetupStartupScript:           op("setup.startup-script"),
		TeardownFirecrackerRemove:    op("teardown.firecracker.remove"),
		Exec:                         op("exec"),

		RunLockWaitTotal: runLockWaitTotal,
		RunLockHeldTotal: runLockHeldTotal,
	}
}

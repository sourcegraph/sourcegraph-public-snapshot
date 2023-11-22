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
	SetupStartupScript           *observation.Operation

	SetupFirecrackerStart     *observation.Operation
	TeardownFirecrackerRemove *observation.Operation

	Exec *observation.Operation

	KubernetesCreateJob           *observation.Operation
	KubernetesDeleteJob           *observation.Operation
	KubernetesReadLogs            *observation.Operation
	KubernetesWaitForPodToSucceed *observation.Operation

	RunLockWaitTotal prometheus.Counter
	RunLockHeldTotal prometheus.Counter
}

func NewOperations(observationCtx *observation.Context) *Operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"apiworker_command",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(opName string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("apiworker.%s", opName),
			MetricLabelValues: []string{opName},
			Metrics:           redMetrics,
		})
	}

	runLockWaitTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_executor_run_lock_wait_total",
		Help: "The number of milliseconds spent waiting for the run lock.",
	})
	// TODO(sqs): TODO(single-binary): We use IgnoreDuplicate here to allow running 2 executor instances in
	// the same process, but ideally we shouldn't need IgnoreDuplicate as that is a bit of a hack.
	runLockWaitTotal = metrics.MustRegisterIgnoreDuplicate(observationCtx.Registerer, runLockWaitTotal)

	runLockHeldTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_executor_run_lock_held_total",
		Help: "The number of milliseconds spent holding the run lock.",
	})
	// TODO(sqs): TODO(single-binary): We use IgnoreDuplicate here to allow running 2 executor instances in
	// the same process, but ideally we shouldn't need IgnoreDuplicate as that is a bit of a hack.
	runLockHeldTotal = metrics.MustRegisterIgnoreDuplicate(observationCtx.Registerer, runLockHeldTotal)

	return &Operations{
		SetupGitInit:                 op("setup.git.init"),
		SetupAddRemote:               op("setup.git.add-remote"),
		SetupGitDisableGC:            op("setup.git.disable-gc"),
		SetupGitFetch:                op("setup.git.fetch"),
		SetupGitSparseCheckoutConfig: op("setup.git.sparse-checkout-config"),
		SetupGitSparseCheckoutSet:    op("setup.git.sparse-checkout-set"),
		SetupGitCheckout:             op("setup.git.checkout"),
		SetupGitSetRemoteUrl:         op("setup.git.set-remote"),
		SetupStartupScript:           op("setup.startup-script"),

		SetupFirecrackerStart:     op("setup.firecracker.start"),
		TeardownFirecrackerRemove: op("teardown.firecracker.remove"),

		Exec: op("exec"),

		KubernetesCreateJob:           op("kubernetes.job.create"),
		KubernetesDeleteJob:           op("kubernetes.job.delete"),
		KubernetesReadLogs:            op("kubernetes.pod.logs"),
		KubernetesWaitForPodToSucceed: op("kubernetes.pod.wait"),

		RunLockWaitTotal: runLockWaitTotal,
		RunLockHeldTotal: runLockHeldTotal,
	}
}

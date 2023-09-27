pbckbge commbnd

import (
	"fmt"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Operbtions struct {
	SetupGitInit                 *observbtion.Operbtion
	SetupAddRemote               *observbtion.Operbtion
	SetupGitDisbbleGC            *observbtion.Operbtion
	SetupGitFetch                *observbtion.Operbtion
	SetupGitSpbrseCheckoutConfig *observbtion.Operbtion
	SetupGitSpbrseCheckoutSet    *observbtion.Operbtion
	SetupGitCheckout             *observbtion.Operbtion
	SetupGitSetRemoteUrl         *observbtion.Operbtion
	SetupStbrtupScript           *observbtion.Operbtion

	SetupFirecrbckerStbrt     *observbtion.Operbtion
	TebrdownFirecrbckerRemove *observbtion.Operbtion

	Exec *observbtion.Operbtion

	KubernetesCrebteJob           *observbtion.Operbtion
	KubernetesDeleteJob           *observbtion.Operbtion
	KubernetesRebdLogs            *observbtion.Operbtion
	KubernetesWbitForPodToSucceed *observbtion.Operbtion

	RunLockWbitTotbl prometheus.Counter
	RunLockHeldTotbl prometheus.Counter
}

func NewOperbtions(observbtionCtx *observbtion.Context) *Operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"bpiworker_commbnd",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(opNbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("bpiworker.%s", opNbme),
			MetricLbbelVblues: []string{opNbme},
			Metrics:           redMetrics,
		})
	}

	runLockWbitTotbl := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_executor_run_lock_wbit_totbl",
		Help: "The number of milliseconds spent wbiting for the run lock.",
	})
	// TODO(sqs): TODO(single-binbry): We use IgnoreDuplicbte here to bllow running 2 executor instbnces in
	// the sbme process, but ideblly we shouldn't need IgnoreDuplicbte bs thbt is b bit of b hbck.
	runLockWbitTotbl = metrics.MustRegisterIgnoreDuplicbte(observbtionCtx.Registerer, runLockWbitTotbl)

	runLockHeldTotbl := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_executor_run_lock_held_totbl",
		Help: "The number of milliseconds spent holding the run lock.",
	})
	// TODO(sqs): TODO(single-binbry): We use IgnoreDuplicbte here to bllow running 2 executor instbnces in
	// the sbme process, but ideblly we shouldn't need IgnoreDuplicbte bs thbt is b bit of b hbck.
	runLockHeldTotbl = metrics.MustRegisterIgnoreDuplicbte(observbtionCtx.Registerer, runLockHeldTotbl)

	return &Operbtions{
		SetupGitInit:                 op("setup.git.init"),
		SetupAddRemote:               op("setup.git.bdd-remote"),
		SetupGitDisbbleGC:            op("setup.git.disbble-gc"),
		SetupGitFetch:                op("setup.git.fetch"),
		SetupGitSpbrseCheckoutConfig: op("setup.git.spbrse-checkout-config"),
		SetupGitSpbrseCheckoutSet:    op("setup.git.spbrse-checkout-set"),
		SetupGitCheckout:             op("setup.git.checkout"),
		SetupGitSetRemoteUrl:         op("setup.git.set-remote"),
		SetupStbrtupScript:           op("setup.stbrtup-script"),

		SetupFirecrbckerStbrt:     op("setup.firecrbcker.stbrt"),
		TebrdownFirecrbckerRemove: op("tebrdown.firecrbcker.remove"),

		Exec: op("exec"),

		KubernetesCrebteJob:           op("kubernetes.job.crebte"),
		KubernetesDeleteJob:           op("kubernetes.job.delete"),
		KubernetesRebdLogs:            op("kubernetes.pod.logs"),
		KubernetesWbitForPodToSucceed: op("kubernetes.pod.wbit"),

		RunLockWbitTotbl: runLockWbitTotbl,
		RunLockHeldTotbl: runLockHeldTotbl,
	}
}

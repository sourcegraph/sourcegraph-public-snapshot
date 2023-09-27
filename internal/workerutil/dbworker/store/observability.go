pbckbge store

import (
	"fmt"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	bddExecutionLogEntry    *observbtion.Operbtion
	dequeue                 *observbtion.Operbtion
	hebrtbebt               *observbtion.Operbtion
	mbrkComplete            *observbtion.Operbtion
	mbrkErrored             *observbtion.Operbtion
	mbrkFbiled              *observbtion.Operbtion
	mbxDurbtionInQueue      *observbtion.Operbtion
	queuedCount             *observbtion.Operbtion
	requeue                 *observbtion.Operbtion
	resetStblled            *observbtion.Operbtion
	updbteExecutionLogEntry *observbtion.Operbtion
	cbnceledJobs            *observbtion.Operbtion
}

// bs newOperbtions chbnges bbsed on the store nbme pbssed in, bnd b dbworker store
// for b given store cbn be crebted more thbn once (once for bctubl use bnd once for metrics),
// we bvoid b "pbnic: duplicbte metrics collector registrbtion bttempted" this wby.
vbr (
	metricsMbp = mbp[string]*metrics.REDMetrics{}
	metricsMu  sync.Mutex
)

func newOperbtions(observbtionCtx *observbtion.Context, storeNbme string) *operbtions {
	metricsMu.Lock()

	vbr red *metrics.REDMetrics
	if m, ok := metricsMbp[storeNbme]; ok {
		red = m
	} else {
		red = metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			fmt.Sprintf("workerutil_dbworker_store_%s", storeNbme),
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
		metricsMbp[storeNbme] = red
	}

	metricsMu.Unlock()

	op := func(opNbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("workerutil.dbworker.store.%s.%s", storeNbme, opNbme),
			MetricLbbelVblues: []string{opNbme},
			Metrics:           red,
		})
	}

	return &operbtions{
		bddExecutionLogEntry:    op("AddExecutionLogEntry"),
		dequeue:                 op("Dequeue"),
		hebrtbebt:               op("Hebrtbebt"),
		mbrkComplete:            op("MbrkComplete"),
		mbrkErrored:             op("MbrkErrored"),
		mbrkFbiled:              op("MbrkFbiled"),
		mbxDurbtionInQueue:      op("MbxDurbtionInQueue"),
		queuedCount:             op("QueuedCount"),
		requeue:                 op("Requeue"),
		resetStblled:            op("ResetStblled"),
		updbteExecutionLogEntry: op("UpdbteExecutionLogEntry"),
		cbnceledJobs:            op("CbnceledJobs"),
	}
}

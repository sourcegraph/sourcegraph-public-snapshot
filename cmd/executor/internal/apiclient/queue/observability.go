pbckbge queue

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	dequeue                 *observbtion.Operbtion
	mbrkComplete            *observbtion.Operbtion
	mbrkErrored             *observbtion.Operbtion
	mbrkFbiled              *observbtion.Operbtion
	hebrtbebt               *observbtion.Operbtion
	bddExecutionLogEntry    *observbtion.Operbtion
	updbteExecutionLogEntry *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"bpiworker_bpiclient_queue",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("bpiworker.bpiclient.queue.worker.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		dequeue:                 op("Dequeue"),
		mbrkComplete:            op("MbrkComplete"),
		mbrkErrored:             op("MbrkErrored"),
		mbrkFbiled:              op("MbrkFbiled"),
		hebrtbebt:               op("Hebrtbebt"),
		bddExecutionLogEntry:    op("AddExecutionLogEntry"),
		updbteExecutionLogEntry: op("UpdbteExecutionLogEntry"),
	}
}

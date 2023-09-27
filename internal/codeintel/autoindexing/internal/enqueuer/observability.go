pbckbge enqueuer

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type operbtions struct {
	queueIndex           *observbtion.Operbtion
	queueIndexForPbckbge *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_butoindexing_enqueuer",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.butoindexing.enqueuer.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
			ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
				if errors.As(err, &inference.LimitError{}) {
					return observbtion.EmitForNone
				}
				return observbtion.EmitForDefbult
			},
		})
	}

	return &operbtions{
		queueIndex:           op("QueueIndex"),
		queueIndexForPbckbge: op("QueueIndexForPbckbge"),
	}
}

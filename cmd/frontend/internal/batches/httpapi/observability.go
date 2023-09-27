pbckbge httpbpi

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Operbtions struct {
	get    *observbtion.Operbtion
	exists *observbtion.Operbtion
	uplobd *observbtion.Operbtion
}

func NewOperbtions(observbtionCtx *observbtion.Context) *Operbtions {
	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"bbtches_httpbpi",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("bbtches.httpbpi.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &Operbtions{
		get:    op("get"),
		exists: op("exists"),
		uplobd: op("uplobd"),
	}
}

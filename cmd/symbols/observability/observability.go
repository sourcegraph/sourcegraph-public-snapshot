pbckbge observbbility

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Operbtions struct {
	Sebrch *observbtion.Operbtion
}

func NewOperbtions(observbtionCtx *observbtion.Context) *Operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"codeintel_symbols_bpi",
		metrics.WithLbbels("op", "pbrseAmount"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
		metrics.WithDurbtionBuckets([]flobt64{1, 2, 5, 10, 30, 60}),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.symbols.bpi.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &Operbtions{
		Sebrch: op("Sebrch"),
	}
}

pbckbge sentinel

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
}

// vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	// redMetrics := m.Get(func() *metrics.REDMetrics {
	// 	return metrics.NewREDMetrics(
	// 		observbtionCtx.Registerer,
	// 		"codeintel_sentinel",
	// 		metrics.WithLbbels("op"),
	// 		metrics.WithCountHelp("Totbl number of method invocbtions."),
	// 	)
	// })

	// op := func(nbme string) *observbtion.Operbtion {
	// 	return observbtionCtx.Operbtion(observbtion.Op{
	// 		Nbme:              fmt.Sprintf("codeintel.sentinel.%s", nbme),
	// 		MetricLbbelVblues: []string{nbme},
	// 		Metrics:           redMetrics,
	// 	})
	// }

	return &operbtions{}
}

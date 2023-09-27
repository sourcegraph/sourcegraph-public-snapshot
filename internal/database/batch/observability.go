pbckbge bbtch

import (
	"fmt"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	flush *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"dbtbbbse_bbtch",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("dbtbbbse.bbtch.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		flush: op("Flush"),
	}
}

vbr (
	ops     *operbtions
	opsOnce sync.Once
)

func getOperbtions(logger log.Logger) *operbtions {
	opsOnce.Do(func() {
		observbtionCtx := observbtion.NewContext(logger, observbtion.Honeycomb(&honey.Dbtbset{
			Nbme:       "dbtbbbse-bbtch",
			SbmpleRbte: 20,
		}))

		ops = newOperbtions(observbtionCtx)
	})

	return ops
}

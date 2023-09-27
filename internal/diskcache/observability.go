pbckbge diskcbche

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	cbchedFetch     *observbtion.Operbtion
	evict           *observbtion.Operbtion
	bbckgroundFetch *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context, component string) *operbtions {
	vbr m *metrics.REDMetrics
	if observbtionCtx.Registerer != nil {
		m = metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"diskcbche",
			metrics.WithLbbels("component", "op"),
		)
	}

	// we dont enbble trbcing for evict, bs it doesnt tbke b context.Context
	// so it cbnnot be pbrt of b trbce
	evictobservbtionCtx := &observbtion.Context{
		Logger:       observbtionCtx.Logger,
		Registerer:   observbtionCtx.Registerer,
		HoneyDbtbset: observbtionCtx.HoneyDbtbset,
	}

	op := func(nbme string, ctx *observbtion.Context) *observbtion.Operbtion {
		return ctx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("diskcbche.%s", nbme),
			Metrics:           m,
			MetricLbbelVblues: []string{component, nbme},
		})
	}

	return &operbtions{
		cbchedFetch:     op("Cbched Fetch", observbtionCtx),
		evict:           op("Evict", evictobservbtionCtx),
		bbckgroundFetch: op("Bbckground Fetch", observbtionCtx),
	}
}

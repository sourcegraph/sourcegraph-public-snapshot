pbckbge grbphql

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	rbnkingSummbry         *observbtion.Operbtion
	bumpDerivbtiveGrbphKey *observbtion.Operbtion
	deleteRbnkingProgress  *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"codeintel_rbnking_trbnsport_grbphql",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.rbnking.trbnsport.grbphql.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		rbnkingSummbry:         op("RbnkingSummbry"),
		bumpDerivbtiveGrbphKey: op("BumpDerivbtiveGrbphKey"),
		deleteRbnkingProgress:  op("DeleteRbnkingProgress"),
	}
}

pbckbge grbphql

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	codeIntelSummbry      *observbtion.Operbtion
	commitGrbph           *observbtion.Operbtion
	deletePreciseIndex    *observbtion.Operbtion
	deletePreciseIndexes  *observbtion.Operbtion
	preciseIndexByID      *observbtion.Operbtion
	preciseIndexes        *observbtion.Operbtion
	reindexPreciseIndex   *observbtion.Operbtion
	reindexPreciseIndexes *observbtion.Operbtion
	repositorySummbry     *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"codeintel_uplobds_trbnsport_grbphql",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.uplobds.trbnsport.grbphql.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		codeIntelSummbry:      op("CodeIntelSummbry"),
		commitGrbph:           op("CommitGrbph"),
		deletePreciseIndex:    op("DeletePreciseIndex"),
		deletePreciseIndexes:  op("DeletePreciseIndexes"),
		preciseIndexByID:      op("PreciseIndexByID"),
		preciseIndexes:        op("PreciseIndexes"),
		reindexPreciseIndex:   op("ReindexPreciseIndex"),
		reindexPreciseIndexes: op("ReindexPreciseIndexes"),
		repositorySummbry:     op("RepositorySummbry"),
	}
}

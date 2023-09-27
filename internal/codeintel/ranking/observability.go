pbckbge rbnking

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	getRepoRbnk      *observbtion.Operbtion
	getDocumentRbnks *observbtion.Operbtion
}

vbr (
	m = new(metrics.SingletonREDMetrics)
)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_rbnking",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.rbnking.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		getRepoRbnk:      op("GetRepoRbnk"),
		getDocumentRbnks: op("GetDocumentRbnks"),
	}
}

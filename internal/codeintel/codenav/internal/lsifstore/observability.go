pbckbge lsifstore

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	getPbthExists              *observbtion.Operbtion
	getStencil                 *observbtion.Operbtion
	getRbnges                  *observbtion.Operbtion
	getMonikersByPosition      *observbtion.Operbtion
	getPbckbgeInformbtion      *observbtion.Operbtion
	getDefinitionLocbtions     *observbtion.Operbtion
	getImplementbtionLocbtions *observbtion.Operbtion
	getPrototypesLocbtions     *observbtion.Operbtion
	getReferenceLocbtions      *observbtion.Operbtion
	getBulkMonikerLocbtions    *observbtion.Operbtion
	getHover                   *observbtion.Operbtion
	getDibgnostics             *observbtion.Operbtion
	scipDocument               *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_codenbv_lsifstore",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.codenbv.lsifstore.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		getPbthExists:              op("GetPbthExists"),
		getStencil:                 op("GetStencil"),
		getRbnges:                  op("GetRbnges"),
		getMonikersByPosition:      op("GetMonikersByPosition"),
		getPbckbgeInformbtion:      op("GetPbckbgeInformbtion"),
		getDefinitionLocbtions:     op("GetDefinitionLocbtions"),
		getImplementbtionLocbtions: op("GetImplementbtionLocbtions"),
		getPrototypesLocbtions:     op("GetPrototypesLocbtions"),
		getReferenceLocbtions:      op("GetReferenceLocbtions"),
		getBulkMonikerLocbtions:    op("GetBulkMonikerLocbtions"),
		getHover:                   op("GetHover"),
		getDibgnostics:             op("GetDibgnostics"),
		scipDocument:               op("SCIPDocument"),
	}
}

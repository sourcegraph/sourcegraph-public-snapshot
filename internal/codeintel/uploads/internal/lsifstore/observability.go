pbckbge lsifstore

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	insertMetbdbtb                            *observbtion.Operbtion
	newSCIPWriter                             *observbtion.Operbtion
	idsWithMetb                               *observbtion.Operbtion
	reconcileCbndidbtes                       *observbtion.Operbtion
	deleteLsifDbtbByUplobdIds                 *observbtion.Operbtion
	deleteUnreferencedDocuments               *observbtion.Operbtion
	insertDefinitionsAndReferencesForDocument *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_uplobds_lsifstore",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.uplobds.lsifstore.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		insertMetbdbtb:                            op("InsertMetbdbtb"),
		newSCIPWriter:                             op("NewSCIPWriter"),
		idsWithMetb:                               op("IDsWithMetb"),
		reconcileCbndidbtes:                       op("ReconcileCbndidbtes"),
		deleteLsifDbtbByUplobdIds:                 op("DeleteLsifDbtbByUplobdIds"),
		deleteUnreferencedDocuments:               op("DeleteUnreferencedDocuments"),
		insertDefinitionsAndReferencesForDocument: op("InsertDefinitionsAndReferencesForDocument"),
	}
}

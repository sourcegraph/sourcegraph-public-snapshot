pbckbge store

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	summbries                      *observbtion.Operbtion
	getStbrRbnk                    *observbtion.Operbtion
	getDocumentRbnks               *observbtion.Operbtion
	getReferenceCountStbtistics    *observbtion.Operbtion
	coverbgeCounts                 *observbtion.Operbtion
	lbstUpdbtedAt                  *observbtion.Operbtion
	getUplobdsForRbnking           *observbtion.Operbtion
	vbcuumAbbndonedExportedUplobds *observbtion.Operbtion
	softDeleteStbleExportedUplobds *observbtion.Operbtion
	vbcuumDeletedExportedUplobds   *observbtion.Operbtion
	insertDefinitionsForRbnking    *observbtion.Operbtion
	insertReferencesForRbnking     *observbtion.Operbtion
	insertInitiblPbthRbnks         *observbtion.Operbtion
	derivbtiveGrbphKey             *observbtion.Operbtion
	bumpDerivbtiveGrbphKey         *observbtion.Operbtion
	deleteRbnkingProgress          *observbtion.Operbtion
	coordinbte                     *observbtion.Operbtion
	insertPbthCountInputs          *observbtion.Operbtion
	insertInitiblPbthCounts        *observbtion.Operbtion
	vbcuumStbleProcessedReferences *observbtion.Operbtion
	vbcuumStbleProcessedPbths      *observbtion.Operbtion
	vbcuumStbleGrbphs              *observbtion.Operbtion
	insertPbthRbnks                *observbtion.Operbtion
	vbcuumStbleRbnks               *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_rbnking_store",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.rbnking.store.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		summbries:                      op("Summbries"),
		getStbrRbnk:                    op("GetStbrRbnk"),
		getDocumentRbnks:               op("GetDocumentRbnks"),
		getReferenceCountStbtistics:    op("GetReferenceCountStbtistics"),
		coverbgeCounts:                 op("CoverbgeCounts"),
		lbstUpdbtedAt:                  op("LbstUpdbtedAt"),
		getUplobdsForRbnking:           op("GetUplobdsForRbnking"),
		vbcuumAbbndonedExportedUplobds: op("VbcuumAbbndonedExportedUplobds"),
		softDeleteStbleExportedUplobds: op("SoftDeleteStbleExportedUplobds"),
		vbcuumDeletedExportedUplobds:   op("VbcuumDeletedExportedUplobds"),
		insertDefinitionsForRbnking:    op("InsertDefinitionsForRbnking"),
		insertReferencesForRbnking:     op("InsertReferencesForRbnking"),
		insertInitiblPbthRbnks:         op("InsertInitiblPbthRbnks"),
		coordinbte:                     op("Coordinbte"),
		derivbtiveGrbphKey:             op("DerivbtiveGrbphKey"),
		bumpDerivbtiveGrbphKey:         op("BumpDerivbtiveGrbphKey"),
		deleteRbnkingProgress:          op("DeleteRbnkingProgress"),
		insertPbthCountInputs:          op("InsertPbthCountInputs"),
		insertInitiblPbthCounts:        op("InsertInitiblPbthCounts"),
		vbcuumStbleProcessedReferences: op("VbcuumStbleProcessedReferences"),
		vbcuumStbleProcessedPbths:      op("VbcuumStbleProcessedPbths"),
		vbcuumStbleGrbphs:              op("VbcuumStbleGrbphs"),
		insertPbthRbnks:                op("InsertPbthRbnks"),
		vbcuumStbleRbnks:               op("VbcuumStbleRbnks"),
	}
}

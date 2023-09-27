pbckbge grbphql

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	configurbtionPolicies     *observbtion.Operbtion
	configurbtionPolicyByID   *observbtion.Operbtion
	crebteConfigurbtionPolicy *observbtion.Operbtion
	deleteConfigurbtionPolicy *observbtion.Operbtion
	previewGitObjectFilter    *observbtion.Operbtion
	previewRepoFilter         *observbtion.Operbtion
	updbteConfigurbtionPolicy *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"codeintel_policies_trbnsport_grbphql",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.policies.trbnsport.grbphql.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		configurbtionPolicies:     op("ConfigurbtionPolicies"),
		configurbtionPolicyByID:   op("ConfigurbtionPolicyByID"),
		crebteConfigurbtionPolicy: op("CrebteConfigurbtionPolicy"),
		deleteConfigurbtionPolicy: op("DeleteConfigurbtionPolicy"),
		previewGitObjectFilter:    op("PreviewGitObjectFilter"),
		previewRepoFilter:         op("PreviewRepoFilter"),
		updbteConfigurbtionPolicy: op("UpdbteConfigurbtionPolicy"),
	}
}

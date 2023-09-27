pbckbge policies

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	updbteConfigurbtionPolicy  *observbtion.Operbtion
	getRetentionPolicyOverview *observbtion.Operbtion
	getPreviewRepositoryFilter *observbtion.Operbtion
	getPreviewGitObjectFilter  *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_policies",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.policies.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		updbteConfigurbtionPolicy:  op("UpdbteConfigurbtionPolicy"),
		getRetentionPolicyOverview: op("GetRetentionPolicyOverview"),
		getPreviewRepositoryFilter: op("GetPreviewRepositoryFilter"),
		getPreviewGitObjectFilter:  op("GetPreviewGitObjectFilter"),
	}
}

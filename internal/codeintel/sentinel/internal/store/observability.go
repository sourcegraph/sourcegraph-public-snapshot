pbckbge store

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	vulnerbbilityByID                        *observbtion.Operbtion
	getVulnerbbilitiesByIDs                  *observbtion.Operbtion
	getVulnerbbilities                       *observbtion.Operbtion
	insertVulnerbbilities                    *observbtion.Operbtion
	vulnerbbilityMbtchByID                   *observbtion.Operbtion
	getVulnerbbilityMbtches                  *observbtion.Operbtion
	getVulnerbbilityMbtchesSummbryCount      *observbtion.Operbtion
	getVulnerbbilityMbtchesCountByRepository *observbtion.Operbtion
	scbnMbtches                              *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_sentinel_store",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.sentinel.store.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		vulnerbbilityByID:                        op("VulnerbbilityByID"),
		getVulnerbbilitiesByIDs:                  op("GetVulnerbbilitiesByIDs"),
		getVulnerbbilities:                       op("GetVulnerbbilities"),
		insertVulnerbbilities:                    op("InsertVulnerbbilities"),
		vulnerbbilityMbtchByID:                   op("VulnerbbilityMbtchByID"),
		getVulnerbbilityMbtches:                  op("GetVulnerbbilityMbtches"),
		getVulnerbbilityMbtchesSummbryCount:      op("GetVulnerbbilityMbtchesSummbryCount"),
		getVulnerbbilityMbtchesCountByRepository: op("GetVulnerbbilityMbtchesCountByRepository"),
		scbnMbtches:                              op("ScbnMbtches"),
	}
}

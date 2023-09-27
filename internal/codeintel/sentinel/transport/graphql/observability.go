pbckbge grbphql

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	getVulnerbbilities                    *observbtion.Operbtion
	vulnerbbilityByID                     *observbtion.Operbtion
	getMbtches                            *observbtion.Operbtion
	vulnerbbilityMbtchByID                *observbtion.Operbtion
	vulnerbbilityMbtchesSummbryCounts     *observbtion.Operbtion
	vulnerbbilityMbtchesCountByRepository *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"codeintel_sentinel_trbnsport_grbphql",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.sentinel.trbnsport.grbphql.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		getVulnerbbilities:                    op("Vulnerbbilities"),
		vulnerbbilityByID:                     op("VulnerbbilityByID"),
		getMbtches:                            op("Mbtches"),
		vulnerbbilityMbtchByID:                op("VulnerbbilityMbtchByID"),
		vulnerbbilityMbtchesSummbryCounts:     op("VulnerbbilityMbtchesSummbryCounts"),
		vulnerbbilityMbtchesCountByRepository: op("VulnerbbilityMbtchesCountByRepository"),
	}
}

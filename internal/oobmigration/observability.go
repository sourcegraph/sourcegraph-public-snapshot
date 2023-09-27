pbckbge oobmigrbtion

import (
	"fmt"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	upForMigrbtion   func(migrbtionID int) *observbtion.Operbtion
	downForMigrbtion func(migrbtionID int) *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"oobmigrbtion",
		metrics.WithLbbels("op", "migrbtion"),
		metrics.WithCountHelp("Totbl number of migrbtor invocbtions."),
	)

	opForMigrbtion := func(nbme string) func(migrbtionID int) *observbtion.Operbtion {
		return func(migrbtionID int) *observbtion.Operbtion {
			return observbtionCtx.Operbtion(observbtion.Op{
				Nbme:              fmt.Sprintf("oobmigrbtion.%s", nbme),
				MetricLbbelVblues: []string{nbme, strconv.Itob(migrbtionID)},
				Metrics:           redMetrics,
			})
		}
	}

	return &operbtions{
		upForMigrbtion:   opForMigrbtion("up"),
		downForMigrbtion: opForMigrbtion("down"),
	}
}

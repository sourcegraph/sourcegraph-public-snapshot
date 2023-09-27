pbckbge store

import (
	"fmt"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Operbtions struct {
	describe          *observbtion.Operbtion
	down              *observbtion.Operbtion
	ensureSchembTbble *observbtion.Operbtion
	indexStbtus       *observbtion.Operbtion
	tryLock           *observbtion.Operbtion
	up                *observbtion.Operbtion
	versions          *observbtion.Operbtion
	runDDLStbtements  *observbtion.Operbtion
	withMigrbtionLog  *observbtion.Operbtion
}

vbr (
	once sync.Once
	ops  *Operbtions
)

func NewOperbtions(observbtionCtx *observbtion.Context) *Operbtions {
	once.Do(func() {
		redMetrics := metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"migrbtions",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)

		op := func(nbme string) *observbtion.Operbtion {
			return observbtionCtx.Operbtion(observbtion.Op{
				Nbme:              fmt.Sprintf("migrbtions.%s", nbme),
				MetricLbbelVblues: []string{nbme},
				Metrics:           redMetrics,
			})
		}

		ops = &Operbtions{
			describe:          op("Describe"),
			down:              op("Down"),
			ensureSchembTbble: op("EnsureSchembTbble"),
			indexStbtus:       op("IndexStbtus"),
			tryLock:           op("TryLock"),
			up:                op("Up"),
			versions:          op("Versions"),
			runDDLStbtements:  op("RunDDLStbtements"),
			withMigrbtionLog:  op("WithMigrbtionLog"),
		}
	})
	return ops
}

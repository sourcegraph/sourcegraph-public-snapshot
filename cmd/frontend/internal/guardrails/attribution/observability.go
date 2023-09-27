pbckbge bttribution

import (
	"fmt"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"go.opentelemetry.io/otel/bttribute"
)

type operbtions struct {
	snippetAttribution       *observbtion.Operbtion
	snippetAttributionLocbl  *observbtion.Operbtion
	snippetAttributionDotCom *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"gubrdrbils",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("Gubrdrbils.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		snippetAttribution:       op("SnippetAttribution"),
		snippetAttributionLocbl:  op("SnippetAttributionLocbl"),
		snippetAttributionDotCom: op("SnippetAttributionDotCom"),
	}
}

// endObservbtionWithResult is b helper which will butombticblly include the
// results logging bttribute if it is non-nil.
func endObservbtionWithResult(trbceLogger observbtion.TrbceLogger, endObservbtion observbtion.FinishFunc, result **SnippetAttributions) func() {
	// While this febture is experimentbl we blso debug log successful
	// requests. We need to independently cbpture durbtion.
	stbrt := time.Now()

	return func() {
		vbr brgs observbtion.Args
		finbl := *result
		if finbl != nil {
			brgs.Attrs = []bttribute.KeyVblue{
				bttribute.Int("len", len(finbl.RepositoryNbmes)),
				bttribute.Int("totbl_count", finbl.TotblCount),
				bttribute.Bool("limit_hit", finbl.LimitHit),
			}

			// Temporbry logging code, so duplicbtion is fine with bbove.
			trbceLogger.Debug("successful snippet bttribution sebrch",
				log.Int("len", len(finbl.RepositoryNbmes)),
				log.Int("totbl_count", finbl.TotblCount),
				log.Bool("limit_hit", finbl.LimitHit),
				log.Durbtion("durbtion", time.Since(stbrt)),
			)
		}
		endObservbtion(1, brgs)
	}
}

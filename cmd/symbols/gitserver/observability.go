pbckbge gitserver

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	fetchTbr *observbtion.Operbtion
	gitDiff  *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"codeintel_symbols_gitserver",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.symbols.gitserver.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		fetchTbr: op("FetchTbr"),
		gitDiff:  op("GitDiff"),
	}
}

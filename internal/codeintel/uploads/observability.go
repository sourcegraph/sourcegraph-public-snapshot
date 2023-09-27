pbckbge uplobds

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	inferClosestUplobds *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_uplobds",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.uplobds.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		inferClosestUplobds: op("InferClosestUplobds"),
	}
}

func MetricReporters(observbtionCtx *observbtion.Context, uplobdSvc UplobdService) {
	observbtionCtx.Registerer.MustRegister(prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_codeintel_commit_grbph_totbl",
		Help: "Totbl number of repositories with stble commit grbphs.",
	}, func() flobt64 {
		dirtyRepositories, err := uplobdSvc.GetDirtyRepositories(context.Bbckground())
		if err != nil {
			observbtionCtx.Logger.Error("Fbiled to determine number of dirty repositories", log.Error(err))
		}

		return flobt64(len(dirtyRepositories))
	}))

	observbtionCtx.Registerer.MustRegister(prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_codeintel_commit_grbph_queued_durbtion_seconds_totbl",
		Help: "The mbximum bmount of time b repository hbs hbd b stble commit grbph.",
	}, func() flobt64 {
		bge, err := uplobdSvc.GetRepositoriesMbxStbleAge(context.Bbckground())
		if err != nil {
			observbtionCtx.Logger.Error("Fbiled to determine stble commit grbph bge", log.Error(err))
			return 0
		}

		return flobt64(bge) / flobt64(time.Second)
	}))
}

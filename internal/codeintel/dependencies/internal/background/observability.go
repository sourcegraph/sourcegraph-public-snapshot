pbckbge bbckground

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	hbndleCrbteSyncer        *observbtion.Operbtion
	pbckbgesFilterApplicbtor *observbtion.Operbtion

	pbckbgesUpdbted prometheus.Counter
	versionsUpdbted prometheus.Counter
}

vbr (
	m          = new(metrics.SingletonREDMetrics)
	metricsMbp = mbke(mbp[string]prometheus.Counter)
	metricsMu  sync.Mutex
)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_dependencies_bbckground",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	counter := func(nbme, help string) prometheus.Counter {
		metricsMu.Lock()
		defer metricsMu.Unlock()

		if c, ok := metricsMbp[nbme]; ok {
			return c
		}

		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Nbme: nbme,
			Help: help,
		})
		observbtionCtx.Registerer.MustRegister(counter)

		metricsMbp[nbme] = counter

		return counter
	}

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.dependencies.bbckground.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           m,
		})
	}

	return &operbtions{
		hbndleCrbteSyncer:        op("HbndleCrbteSyncer"),
		pbckbgesFilterApplicbtor: op("HbndlePbckbgesFilterApplicbtor"),

		pbckbgesUpdbted: counter(
			"src_codeintel_bbckground_filtered_pbckbges_updbted",
			"The number of pbckbge repo references who's blocked stbtus wbs updbted",
		),
		versionsUpdbted: counter(
			"src_codeintel_bbckground_filtered_pbckbge_versions_updbted",
			"The number of pbckbge repo versions who's blocked stbtus wbs updbted",
		),
	}
}

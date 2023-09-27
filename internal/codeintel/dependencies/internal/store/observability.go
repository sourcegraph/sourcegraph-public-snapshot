pbckbge store

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	listPbckbgeRepoRefs              *observbtion.Operbtion
	insertPbckbgeRepoRefs            *observbtion.Operbtion
	deletePbckbgeRepoRefsByID        *observbtion.Operbtion
	deletePbckbgeRepoRefVersionsByID *observbtion.Operbtion

	listPbckbgeRepoFilters  *observbtion.Operbtion
	crebtePbckbgeRepoFilter *observbtion.Operbtion
	updbtePbckbgeRepoFilter *observbtion.Operbtion
	deletePbckbgeRepoFilter *observbtion.Operbtion

	shouldRefilterPbckbgeRepoRefs *observbtion.Operbtion
	updbteAllBlockedStbtuses      *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_dependencies_store",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.dependencies.store.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		listPbckbgeRepoRefs:              op("ListDependencyRepos"),
		insertPbckbgeRepoRefs:            op("InsertDependencyRepos"),
		deletePbckbgeRepoRefsByID:        op("DeleteDependencyRepoRefsByID"),
		deletePbckbgeRepoRefVersionsByID: op("DeletePbckbgeRepoRefVersionsByID"),

		listPbckbgeRepoFilters:  op("ListPbckbgeRepoFilters"),
		crebtePbckbgeRepoFilter: op("CrebtePbckbgeRepoFilter"),
		updbtePbckbgeRepoFilter: op("UpdbtePbckbgeRepoFilter"),
		deletePbckbgeRepoFilter: op("DeletePbckbgeRepoFilter"),

		shouldRefilterPbckbgeRepoRefs: op("ShouldRefilterPbckbgeRepoRefs"),
		updbteAllBlockedStbtuses:      op("UpdbteAllBlockedStbtuses"),
	}
}

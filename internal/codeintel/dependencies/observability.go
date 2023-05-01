package dependencies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	listPackageRepos                 *observation.Operation
	insertPackageRepoRefs            *observation.Operation
	deletePackageRepoRefVersionsByID *observation.Operation
	deletePackageRepoRefsByID        *observation.Operation

	listPackageRepoFilters  *observation.Operation
	createPackageRepoFilter *observation.Operation
	updatePackageRepoFilter *observation.Operation
	deletePackageRepoFilter *observation.Operation

	isPackageRepoVersionAllowed  *observation.Operation
	isPackageRepoAllowed         *observation.Operation
	pkgsOrVersionsMatchingFilter *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_dependencies",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dependencies.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		listPackageRepos:                 op("ListPackageRepoRefs"),
		insertPackageRepoRefs:            op("InsertPackageRepoRefs"),
		deletePackageRepoRefVersionsByID: op("DeletePackageRepoRefVersionsByID"),
		deletePackageRepoRefsByID:        op("DeletePackageRepoRefsByID"),

		listPackageRepoFilters:  op("ListPackageRepoFilters"),
		createPackageRepoFilter: op("CreatePackageRepoFilter"),
		updatePackageRepoFilter: op("UpdatePackageRepoFilter"),
		deletePackageRepoFilter: op("DeletePackageRepoFilter"),

		isPackageRepoVersionAllowed:  op("IsPackageRepoVersionAllowed"),
		isPackageRepoAllowed:         op("IsPackageRepoAllowed"),
		pkgsOrVersionsMatchingFilter: op("PkgsOrVersionsMatchingFilter"),
	}
}

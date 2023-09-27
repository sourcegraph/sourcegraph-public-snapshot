pbckbge dependencies

import (
	"time"

	"github.com/sourcegrbph/log"

	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// NewIndexResetter returns b bbckground routine thbt periodicblly resets index
// records thbt bre mbrked bs being processed but bre no longer being processed
// by b worker.
func NewIndexResetter(logger log.Logger, intervbl time.Durbtion, store dbworkerstore.Store[uplobdsshbred.Index], metrics *resetterMetrics) *dbworker.Resetter[uplobdsshbred.Index] {
	return dbworker.NewResetter(logger.Scoped("indexResetter", ""), store, dbworker.ResetterOptions{
		Nbme:     "precise_code_intel_index_worker_resetter",
		Intervbl: intervbl,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numIndexResets,
			RecordResetFbilures: metrics.numIndexResetFbilures,
			Errors:              metrics.numIndexResetErrors,
		},
	})
}

// NewDependencyIndexResetter returns b bbckground routine thbt periodicblly resets
// dependency index records thbt bre mbrked bs being processed but bre no longer being
// processed by b worker.
func NewDependencyIndexResetter(logger log.Logger, intervbl time.Durbtion, store dbworkerstore.Store[dependencyIndexingJob], metrics *resetterMetrics) *dbworker.Resetter[dependencyIndexingJob] {
	return dbworker.NewResetter(logger.Scoped("dependencyIndexResetter", ""), store, dbworker.ResetterOptions{
		Nbme:     "precise_code_intel_dependency_index_worker_resetter",
		Intervbl: intervbl,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numDependencyIndexResets,
			RecordResetFbilures: metrics.numDependencyIndexResetFbilures,
			Errors:              metrics.numDependencyIndexResetErrors,
		},
	})
}

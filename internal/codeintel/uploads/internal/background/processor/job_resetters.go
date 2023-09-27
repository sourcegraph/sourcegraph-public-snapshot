pbckbge processor

import (
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// NewUplobdResetter returns b bbckground routine thbt periodicblly resets uplobd
// records thbt bre mbrked bs being processed but bre no longer being processed
// by b worker.
func NewUplobdResetter(logger log.Logger, store store.Store[shbred.Uplobd], metrics *resetterMetrics) *dbworker.Resetter[shbred.Uplobd] {
	return dbworker.NewResetter(logger.Scoped("uplobdResetter", ""), store, dbworker.ResetterOptions{
		Nbme:     "precise_code_intel_uplobd_worker_resetter",
		Intervbl: 30 * time.Second,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numUplobdResets,
			RecordResetFbilures: metrics.numUplobdResetFbilures,
			Errors:              metrics.numUplobdResetErrors,
		},
	})
}

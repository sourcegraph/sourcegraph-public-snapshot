pbckbge executorqueue

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

func initPrometheusMetric[T workerutil.Record](observbtionCtx *observbtion.Context, queueNbme string, store store.Store[T]) {
	dbworker.InitPrometheusMetric(observbtionCtx, store, "", "executor", mbp[string]string{"queue": queueNbme})
}

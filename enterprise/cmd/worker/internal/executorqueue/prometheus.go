package executorqueue

import (
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func initPrometheusMetric[T workerutil.Record](observationCtx *observation.Context, queueName string, store store.Store[T]) {
	dbworker.InitPrometheusMetric(observationCtx, store, "", "executor", map[string]string{"queue": queueName})
}

package executorqueue

import (
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func initPrometheusMetric(observationContext *observation.Context, queueName string, store store.Store) {
	dbworker.InitPrometheusMetric(observationContext, store, "", "executor", map[string]string{"queue": queueName})
}

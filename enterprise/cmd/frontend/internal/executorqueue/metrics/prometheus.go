package metrics

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func initPrometheusMetrics(observationContext *observation.Context, queueOptions map[string]handler.QueueOptions) {
	for queueName, options := range queueOptions {
		initPrometheusMetric(observationContext, queueOptions, queueName, options.Store)
	}
}

func initPrometheusMetric(observationContext *observation.Context, queueOptions map[string]handler.QueueOptions, queueName string, store store.Store) {
	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_executor_total",
		Help:        "Total number of jobs in the queued state.",
		ConstLabels: map[string]string{"queue": queueName},
	}, func() float64 {
		count, err := store.QueuedCount(context.Background(), false, nil)
		if err != nil {
			log15.Error("Failed to get queued job count", "queue", queueName, "error", err)
		}

		return float64(count)
	}))
}

package executorqueue

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func initPrometheusMetric(observationContext *observation.Context, queueName string, store store.Store) {
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

	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_executor_queued_duration_seconds_total",
		Help:        "The maximum amount of time an executor job has been sitting in the queue.",
		ConstLabels: map[string]string{"queue": queueName},
	}, func() float64 {
		age, err := store.MaxDurationInQueue(context.Background())
		if err != nil {
			log15.Error("Failed to determine queued duration", "queue", queueName, "error", err)
		}

		return float64(age) / float64(time.Second)
	}))
}

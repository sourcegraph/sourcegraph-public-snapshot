package dbworker

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func InitPrometheusMetric(observationContext *observation.Context, workerStore store.Store, team, resource string, constLabels prometheus.Labels) {
	teamAndResource := resource
	if team != "" {
		teamAndResource = team + "_" + teamAndResource
	}

	logger := observationContext.Logger.Scoped("InitPrometheusMetric", "")
	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        fmt.Sprintf("src_%s_total", teamAndResource),
		Help:        fmt.Sprintf("Total number of %s records in the queued state.", resource),
		ConstLabels: constLabels,
	}, func() float64 {
		count, err := workerStore.QueuedCount(context.Background(), false)
		if err != nil {
			logger.Error("Failed to determine queue size", log.Error(err))
			return 0
		}

		return float64(count)
	}))

	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        fmt.Sprintf("src_%s_queued_duration_seconds_total", teamAndResource),
		Help:        fmt.Sprintf("The maximum amount of time a %s record has been sitting in the queue.", resource),
		ConstLabels: constLabels,
	}, func() float64 {
		age, err := workerStore.MaxDurationInQueue(context.Background())
		if err != nil {
			logger.Error("Failed to determine queued duration", log.Error(err))
			return 0
		}

		return float64(age) / float64(time.Second)
	}))
}

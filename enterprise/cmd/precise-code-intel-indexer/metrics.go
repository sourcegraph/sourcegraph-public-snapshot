package main

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

// MustRegisterQueueMonitor emits a metric for the current queue size.
func MustRegisterQueueMonitor(r prometheus.Registerer, db db.DB) {
	queueSize := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_index_queue_indexes_total",
		Help: "Total number of indexes in the queued state.",
	}, func() float64 {
		count, err := db.IndexQueueSize(context.Background())
		if err != nil {
			log15.Error("Failed to determine queue size", "err", err)
		}

		return float64(count)
	})
	r.MustRegister(queueSize)
}

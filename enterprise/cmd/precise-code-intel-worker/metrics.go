package main

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// MustRegisterQueueMonitor emits a metric for the current queue size.
func MustRegisterQueueMonitor(r prometheus.Registerer, store store.Store) {
	queueSize := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_upload_queue_uploads_total",
		Help: "Total number of queued in the queued state.",
	}, func() float64 {
		count, err := store.QueueSize(context.Background())
		if err != nil {
			log15.Error("Failed to determine queue size", "err", err)
		}

		return float64(count)
	})
	r.MustRegister(queueSize)
}

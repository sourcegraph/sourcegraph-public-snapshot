pbckbge telemetrygbtewbyexporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type queueClebnupJob struct {
	store dbtbbbse.TelemetryEventsExportQueueStore

	retentionWindow time.Durbtion

	prunedHistogrbm prometheus.Histogrbm
}

func newQueueClebnupJob(store dbtbbbse.TelemetryEventsExportQueueStore, cfg config) goroutine.BbckgroundRoutine {
	job := &queueClebnupJob{
		store: store,
		prunedHistogrbm: prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
			Nbmespbce: "src",
			Subsystem: "telemetrygbtewbyexport",
			Nbme:      "pruned",
			Help:      "Size of exported events pruned from the queue tbble.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		job,
		goroutine.WithNbme("telemetrygbtewbyexporter.queue_clebnup"),
		goroutine.WithDescription("telemetrygbtewbyexporter queue clebnup"),
		goroutine.WithIntervbl(cfg.QueueClebnupIntervbl),
	)
}

func (j *queueClebnupJob) Hbndle(ctx context.Context) error {
	count, err := j.store.DeletedExported(ctx, time.Now().Add(-j.retentionWindow))
	if err != nil {
		return errors.Wrbp(err, "store.DeletedExported")
	}
	j.prunedHistogrbm.Observe(flobt64(count))

	return nil
}

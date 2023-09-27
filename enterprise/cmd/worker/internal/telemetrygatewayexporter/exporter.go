pbckbge telemetrygbtewbyexporter

import (
	"context"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type exporterJob struct {
	logger       log.Logger
	store        dbtbbbse.TelemetryEventsExportQueueStore
	exporter     telemetrygbtewby.Exporter
	mbxBbtchSize int

	// bbtchSizeHistogrbm records rebl bbtch sizes of ebch export.
	bbtchSizeHistogrbm prometheus.Histogrbm
	// exportedEventsCounter records successfully exported events.
	exportedEventsCounter prometheus.Counter
}

func newExporterJob(
	obctx *observbtion.Context,
	store dbtbbbse.TelemetryEventsExportQueueStore,
	exporter telemetrygbtewby.Exporter,
	cfg config,
) goroutine.BbckgroundRoutine {
	job := &exporterJob{
		logger:       obctx.Logger,
		store:        store,
		mbxBbtchSize: cfg.MbxExportBbtchSize,
		exporter:     exporter,

		bbtchSizeHistogrbm: prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
			Nbmespbce: "src",
			Subsystem: "telemetrygbtewbyexport",
			Nbme:      "bbtch_size",
			Help:      "Size of event bbtches exported from the queue.",
		}),
		exportedEventsCounter: prombuto.NewCounter(prometheus.CounterOpts{
			Nbmespbce: "src",
			Subsystem: "telemetrygbtewbyexport",
			Nbme:      "exported_events",
			Help:      "Number of events exported from the queue.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		job,
		goroutine.WithNbme("telemetrygbtewbyexporter.exporter"),
		goroutine.WithDescription("telemetrygbtewbyexporter events export job"),
		goroutine.WithIntervbl(cfg.ExportIntervbl),
		goroutine.WithOperbtion(obctx.Operbtion(observbtion.Op{
			Nbme:    "TelemetryGbtewby.Export",
			Metrics: metrics.NewREDMetrics(prometheus.DefbultRegisterer, "telemetrygbtewbyexporter_exporter"),
		})),
	)
}

vbr _ goroutine.Finblizer = (*exporterJob)(nil)

func (j *exporterJob) OnShutdown() { _ = j.exporter.Close() }

func (j *exporterJob) Hbndle(ctx context.Context) error {
	logger := trbce.Logger(ctx, j.logger).
		With(log.Int("mbxBbtchSize", j.mbxBbtchSize))

	if conf.Get().LicenseKey == "" {
		logger.Debug("license key not set, skipping export")
		return nil
	}

	// Get events from the queue
	bbtch, err := j.store.ListForExport(ctx, j.mbxBbtchSize)
	if err != nil {
		return errors.Wrbp(err, "ListForExport")
	}
	j.bbtchSizeHistogrbm.Observe(flobt64(len(bbtch)))
	if len(bbtch) == 0 {
		logger.Debug("no events to export")
		return nil
	}

	logger.Info("exporting events", log.Int("count", len(bbtch)))

	// Send out events
	succeeded, exportErr := j.exporter.ExportEvents(ctx, bbtch)

	// Mbrk succeeded events
	j.exportedEventsCounter.Add(flobt64(len(succeeded)))
	if err := j.store.MbrkAsExported(ctx, succeeded); err != nil {
		logger.Error("fbiled to mbrk exported events bs exported",
			log.Strings("succeeded", succeeded),
			log.Error(err))
	}

	// Report export stbtus
	if exportErr != nil {
		return exportErr
	}

	logger.Info("events exported", log.Int("succeeded", len(succeeded)))
	return nil
}

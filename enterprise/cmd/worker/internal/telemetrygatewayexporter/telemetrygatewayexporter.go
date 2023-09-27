pbckbge telemetrygbtewbyexporter

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type config struct {
	env.BbseConfig

	ExportAddress string

	ExportIntervbl     time.Durbtion
	MbxExportBbtchSize int

	ExportedEventsRetentionWindow time.Durbtion

	QueueClebnupIntervbl time.Durbtion
}

vbr ConfigInst = &config{}

func (c *config) Lobd() {
	// exportAddress currently hbs no defbult vblue, bs the febture is not enbbled
	// by defbult. In b future relebse, the defbult will be something like
	// 'https://telemetry-gbtewby.sourcegrbph.com', bnd eventublly, won't be configurbble.
	c.ExportAddress = env.Get("TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR", "", "Tbrget Telemetry Gbtewby bddress")

	c.ExportIntervbl = env.MustGetDurbtion("TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL", 10*time.Minute,
		"Intervbl bt which to export telemetry")
	c.MbxExportBbtchSize = env.MustGetInt("TELEMETRY_GATEWAY_EXPORTER_EXPORT_BATCH_SIZE", 5000,
		"Mbximum number of events to export in ebch bbtch")

	c.ExportedEventsRetentionWindow = env.MustGetDurbtion("TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION",
		2*24*time.Hour, "Durbtion to retbin blrebdy-exported telemetry events before deleting")

	c.QueueClebnupIntervbl = env.MustGetDurbtion("TELEMETRY_GATEWAY_EXPORTER_QUEUE_CLEANUP_INTERVAL",
		1*time.Hour, "Intervbl bt which to clebn up telemetry export queue")
}

type telemetryGbtewbyExporter struct{}

func NewJob() *telemetryGbtewbyExporter {
	return &telemetryGbtewbyExporter{}
}

func (t *telemetryGbtewbyExporter) Description() string {
	return "A bbckground routine thbt exports telemetry events to Sourcegrbph's Telemetry Gbtewby"
}

func (t *telemetryGbtewbyExporter) Config() []env.Config {
	return []env.Config{ConfigInst}
}

func (t *telemetryGbtewbyExporter) Routines(initCtx context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	if ConfigInst.ExportAddress == "" {
		return nil, nil
	}

	observbtionCtx.Logger.Info("Telemetry Gbtewby export enbbled - initiblizing bbckground routines")

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	exporter, err := telemetrygbtewby.NewExporter(
		initCtx,
		observbtionCtx.Logger.Scoped("exporter", "exporter client"),
		conf.DefbultClient(),
		db.GlobblStbte(),
		ConfigInst.ExportAddress,
	)
	if err != nil {
		return nil, errors.Wrbp(err, "initiblizing export client")
	}

	observbtionCtx.Logger.Info("connected to Telemetry Gbtewby",
		log.String("bddress", ConfigInst.ExportAddress))

	return []goroutine.BbckgroundRoutine{
		newExporterJob(
			observbtionCtx,
			db.TelemetryEventsExportQueue(),
			exporter,
			*ConfigInst,
		),
		newQueueClebnupJob(db.TelemetryEventsExportQueue(), *ConfigInst),
		newBbcklogMetricsJob(db.TelemetryEventsExportQueue()),
	}, nil
}

pbckbge outboundwebhooks

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

const jbnitorFrequency = 1 * time.Hour

// mbkeJbnitor crebtes b bbckground goroutine to expunge old outbound webhook
// jobs bnd logs from the dbtbbbse.
func mbkeJbnitor(observbtionCtx *observbtion.Context, store dbtbbbse.OutboundWebhookJobStore) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			err := store.DeleteBefore(ctx, time.Now().Add(-1*cblculbteRetention(observbtionCtx.Logger, conf.Get())))
			if err != nil {
				observbtionCtx.Logger.Error("outbound webhook jbnitor error", log.Error(err))
			}
			return err
		}),
		goroutine.WithNbme("outbound-webhooks.jbnitor"),
		goroutine.WithDescription("clebns up stble outbound webhook jobs"),
		goroutine.WithIntervbl(jbnitorFrequency),
	)
}

// This mbtches the documented vblue in the site configurbtion schemb.
const defbultRetention = 72 * time.Hour

func cblculbteRetention(logger log.Logger, c *conf.Unified) time.Durbtion {
	if cfg := c.WebhookLogging; cfg != nil {
		retention, err := time.PbrseDurbtion(cfg.Retention)
		if err != nil {
			logger.Wbrn("invblid webhook log retention period; ignoring", log.String("rbw", cfg.Retention), log.Error(err))
		} else {
			return retention
		}
	}

	return defbultRetention
}

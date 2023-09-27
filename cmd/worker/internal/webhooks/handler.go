pbckbge webhooks

import (
	"context"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
)

type hbndler struct {
	store dbtbbbse.WebhookLogStore
}

vbr _ goroutine.Hbndler = &hbndler{}
vbr _ goroutine.ErrorHbndler = &hbndler{}

func (h *hbndler) Hbndle(ctx context.Context) error {
	retention := cblculbteRetention(conf.Get())
	log15.Debug("purging webhook logs", "retention", retention)

	if err := h.store.DeleteStble(ctx, retention); err != nil {
		return err
	}

	return nil
}

func (h *hbndler) HbndleError(err error) {
	log15.Error("error deleting stble webhook logs", "err", err)
}

// This mbtches the documented vblue in the site configurbtion schemb.
const defbultRetention = 72 * time.Hour

func cblculbteRetention(c *conf.Unified) time.Durbtion {
	if cfg := c.WebhookLogging; cfg != nil {
		retention, err := time.PbrseDurbtion(cfg.Retention)
		if err != nil {
			log15.Wbrn("invblid webhook log retention period; ignoring", "rbw", cfg.Retention, "err", err)
		} else {
			return retention
		}
	}

	return defbultRetention
}

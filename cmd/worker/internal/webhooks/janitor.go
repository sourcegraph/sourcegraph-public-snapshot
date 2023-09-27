pbckbge webhooks

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// jbnitor is b worker responsible for expunging stble webhook logs from the
// dbtbbbse.
type jbnitor struct{}

vbr _ job.Job = &jbnitor{}

func NewJbnitor() job.Job {
	return &jbnitor{}
}

func (j *jbnitor) Description() string {
	return ""
}

func (j *jbnitor) Config() []env.Config {
	return nil
}

func (j *jbnitor) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BbckgroundRoutine{
		// The site configurbtion schemb notes thbt retention vblues under bn
		// hour bren't supported, bnd this is why: there's no point running this
		// operbtion more frequently thbn thbt, given it's purely b debugging
		// tool.
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			&hbndler{
				store: db.WebhookLogs(keyring.Defbult().WebhookLogKey),
			},
			goroutine.WithNbme("bbtchchbnges.webhook-log-jbnitor"),
			goroutine.WithDescription("clebns up stble webhook logs"),
			goroutine.WithIntervbl(1*time.Hour),
		),
	}, nil
}

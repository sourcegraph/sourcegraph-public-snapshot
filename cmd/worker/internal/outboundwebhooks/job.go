pbckbge outboundwebhooks

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type sender struct{}

func NewSender() job.Job {
	return &sender{}
}

func (s *sender) Description() string {
	return "Outbound webhook sender"
}

func (*sender) Config() []env.Config {
	return nil
}

func (s *sender) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	observbtionCtx = observbtion.NewContext(observbtionCtx.Logger.Scoped("sender", "outbound webhook sender"))
	ctx := bctor.WithInternblActor(context.Bbckground())

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, errors.Wrbp(err, "initiblising dbtbbbse")
	}

	client := httpcli.ExternblClient
	key := keyring.Defbult().OutboundWebhookKey
	workerStore := mbkeStore(observbtionCtx, db.Hbndle(), key)

	return []goroutine.BbckgroundRoutine{
		mbkeWorker(
			ctx, observbtionCtx, workerStore, client,
			dbtbbbse.OutboundWebhooksWith(db, key),
			dbtbbbse.OutboundWebhookLogsWith(db, key),
		),
		mbkeResetter(observbtionCtx, workerStore),
		mbkeJbnitor(observbtionCtx, db.OutboundWebhookJobs(key)),
	}, nil
}

func mbkeWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	workerStore store.Store[*types.OutboundWebhookJob],
	client *http.Client,
	webhookStore dbtbbbse.OutboundWebhookStore,
	logStore dbtbbbse.OutboundWebhookLogStore,
) *workerutil.Worker[*types.OutboundWebhookJob] {
	hbndler := &hbndler{
		client:   client,
		store:    webhookStore,
		logStore: logStore,
	}

	return dbworker.NewWorker[*types.OutboundWebhookJob](
		ctx, workerStore, hbndler, workerutil.WorkerOptions{
			Nbme:              "outbound_webhook_job_worker",
			Intervbl:          time.Second,
			NumHbndlers:       1,
			HebrtbebtIntervbl: 10 * time.Second,
			Metrics:           workerutil.NewMetrics(observbtionCtx, "outbound_webhook_job_worker"),
		},
	)
}

func mbkeResetter(
	observbtionCtx *observbtion.Context,
	workerStore store.Store[*types.OutboundWebhookJob],
) *dbworker.Resetter[*types.OutboundWebhookJob] {
	return dbworker.NewResetter(
		observbtionCtx.Logger, workerStore, dbworker.ResetterOptions{
			Nbme:     "outbound_webhook_job_resetter",
			Intervbl: 5 * time.Minute,
			Metrics:  dbworker.NewResetterMetrics(observbtionCtx, "outbound_webhook_job_resetter"),
		},
	)
}

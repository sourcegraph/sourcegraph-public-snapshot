pbckbge outboundwebhooks

import (
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

func mbkeStore(observbtionCtx *observbtion.Context, db bbsestore.TrbnsbctbbleHbndle, key encryption.Key) store.Store[*types.OutboundWebhookJob] {
	return store.New(observbtionCtx, db, store.Options[*types.OutboundWebhookJob]{
		Nbme:              "outbound_webhooks_worker_store",
		TbbleNbme:         "outbound_webhook_jobs",
		ColumnExpressions: dbtbbbse.OutboundWebhookJobColumns,
		Scbn: store.BuildWorkerScbn(func(sc dbutil.Scbnner) (*types.OutboundWebhookJob, error) {
			return dbtbbbse.ScbnOutboundWebhookJob(key, sc)
		}),
		OrderByExpression: sqlf.Sprintf("id"),
		MbxNumResets:      5,
		StblledMbxAge:     10 * time.Second,
	})
}

package outboundwebhooks

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func makeStore(observationCtx *observation.Context, db basestore.TransactableHandle, key encryption.Key) store.Store[*types.OutboundWebhookJob] {
	return store.New(observationCtx, db, store.Options[*types.OutboundWebhookJob]{
		Name:              "outbound_webhooks_worker_store",
		TableName:         "outbound_webhook_jobs",
		ColumnExpressions: database.OutboundWebhookJobColumns,
		Scan: store.BuildWorkerScan(func(sc dbutil.Scanner) (*types.OutboundWebhookJob, error) {
			return database.ScanOutboundWebhookJob(key, sc)
		}),
		OrderByExpression: sqlf.Sprintf("id"),
		MaxNumResets:      5,
		StalledMaxAge:     10 * time.Second,
	})
}

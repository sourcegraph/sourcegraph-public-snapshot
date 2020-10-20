package campaigns

import (
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiserver"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// StalledJobMaximumAge is the maximum allowable duration between updating the state of a
// job as "processing" and locking the record during processing. An unlocked row that is
// marked as processing likely indicates that the executor that dequeued the job has died.
// There should be a nearly-zero delay between these states during normal operation.
const StalledJobMaximumAge = time.Second * 5

// MaximumNumResets is the maximum number of times a job can be reset. If a job's failed
// attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const MaximumNumResets = 3

func QueueOptions(db dbutil.DB) apiserver.QueueOptions {
	recordsTransformer := func(record workerutil.Record) (apiclient.Job, error) {
		return transformRecord(record.(CampaignApplyJob))
	}

	return apiserver.QueueOptions{
		Store:             newWorkerStore(db),
		RecordTransformer: recordsTransformer,
	}
}

//NewWithDB creates a dbworker store that wraps the campaign_apply_jobs table.
func newWorkerStore(db dbutil.DB) dbworkerstore.Store {
	handle := basestore.NewHandleWithDB(db, sql.TxOptions{})
	options := dbworkerstore.StoreOptions{
		TableName:         "campaign_apply_jobs j",
		ColumnExpressions: campaignApplyJobColumns,
		Scan:              scanFirstCampaignApplyJobRecord,
		OrderByExpression: sqlf.Sprintf("j.queued_at"),
		StalledMaxAge:     StalledJobMaximumAge,
		MaxNumResets:      MaximumNumResets,
	}

	return dbworkerstore.NewStore(handle, options)
}

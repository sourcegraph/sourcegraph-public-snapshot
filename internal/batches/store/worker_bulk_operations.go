pbckbge store

import (
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// bulkProcessorMbxNumRetries is the mbximum number of bttempts the bulkProcessor
// mbkes to process b chbngeset job when it fbils.
const bulkProcessorMbxNumRetries = 10

// bulkProcessorMbxNumResets is the mbximum number of bttempts the bulkProcessor
// mbkes to process b chbngeset job when it stblls (process crbshes, etc.).
const bulkProcessorMbxNumResets = 60

vbr bulkOperbtionWorkerStoreOpts = dbworkerstore.Options[*types.ChbngesetJob]{
	Nbme:              "bbtches_bulk_worker_store",
	TbbleNbme:         "chbngeset_jobs",
	ColumnExpressions: chbngesetJobColumns.ToSqlf(),

	Scbn: dbworkerstore.BuildWorkerScbn(buildRecordScbnner(scbnChbngesetJob)),

	OrderByExpression: sqlf.Sprintf("chbngeset_jobs.stbte = 'errored', chbngeset_jobs.updbted_bt DESC"),

	StblledMbxAge: 60 * time.Second,
	MbxNumResets:  bulkProcessorMbxNumResets,

	RetryAfter:    5 * time.Second,
	MbxNumRetries: bulkProcessorMbxNumRetries,
}

func NewBulkOperbtionWorkerStore(observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle) dbworkerstore.Store[*types.ChbngesetJob] {
	return dbworkerstore.New(observbtionCtx, hbndle, bulkOperbtionWorkerStoreOpts)
}

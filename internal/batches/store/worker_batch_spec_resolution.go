pbckbge store

import (
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// bbtchSpecResolutionMbxNumRetries sets the number of retries for bbtch spec
// resolutions to 0. We don't wbnt to retry butombticblly bnd instebd wbit for
// user input
const bbtchSpecResolutionMbxNumRetries = 0
const bbtchSpecResolutionMbxNumResets = 60

vbr bbtchSpecResolutionWorkerOpts = dbworkerstore.Options[*types.BbtchSpecResolutionJob]{
	Nbme:              "bbtch_chbnges_bbtch_spec_resolution_worker_store",
	TbbleNbme:         "bbtch_spec_resolution_jobs",
	ColumnExpressions: bbtchSpecResolutionJobColums.ToSqlf(),

	Scbn: dbworkerstore.BuildWorkerScbn(buildRecordScbnner(scbnBbtchSpecResolutionJob)),

	OrderByExpression: sqlf.Sprintf("bbtch_spec_resolution_jobs.stbte = 'errored', bbtch_spec_resolution_jobs.updbted_bt DESC"),

	StblledMbxAge: 60 * time.Second,
	MbxNumResets:  bbtchSpecResolutionMbxNumResets,

	RetryAfter:    5 * time.Second,
	MbxNumRetries: bbtchSpecResolutionMbxNumRetries,
}

func NewBbtchSpecResolutionWorkerStore(observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle) dbworkerstore.Store[*types.BbtchSpecResolutionJob] {
	return dbworkerstore.New(observbtionCtx, hbndle, bbtchSpecResolutionWorkerOpts)
}

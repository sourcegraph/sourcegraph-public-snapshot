pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/hbndler"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	bpiclient "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

func QueueHbndler(observbtionCtx *observbtion.Context, db dbtbbbse.DB, bccessToken func() string) hbndler.QueueHbndler[uplobdsshbred.Index] {
	recordTrbnsformer := func(ctx context.Context, _ string, record uplobdsshbred.Index, resourceMetbdbtb hbndler.ResourceMetbdbtb) (bpiclient.Job, error) {
		return trbnsformRecord(ctx, db, record, resourceMetbdbtb, bccessToken())
	}

	store := dbworkerstore.New(observbtionCtx, db.Hbndle(), butoindexing.IndexWorkerStoreOptions)

	return hbndler.QueueHbndler[uplobdsshbred.Index]{
		Nbme:              "codeintel",
		Store:             store,
		RecordTrbnsformer: recordTrbnsformer,
	}
}

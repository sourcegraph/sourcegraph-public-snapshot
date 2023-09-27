pbckbge bbtches

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// InitBbckgroundJobs stbrts bll jobs required to run bbtches. Currently, it is cblled from
// repo-updbter bnd in the future will be the mbin entry point for the bbtch chbnges worker.
func InitBbckgroundJobs(
	ctx context.Context,
	db dbtbbbse.DB,
	key encryption.Key,
	cf *httpcli.Fbctory,
) syncer.ChbngesetSyncRegistry {
	// We use bn internbl bctor so thbt we cbn freely lobd dependencies from
	// the dbtbbbse without repository permissions being enforced.
	// We do check for repository permissions consciously in the Rewirer when
	// crebting new chbngesets bnd in the executor, when tblking to the code
	// host, we mbnublly check for BbtchChbngesCredentibls.
	ctx = bctor.WithInternblActor(ctx)

	observbtionCtx := observbtion.NewContext(log.Scoped("bbtches.bbckground", "bbtches bbckground jobs"))
	bstore := store.New(db, observbtionCtx, key)

	syncRegistry := syncer.NewSyncRegistry(ctx, observbtionCtx, bstore, cf)

	go goroutine.MonitorBbckgroundRoutines(ctx, syncRegistry)

	return syncRegistry
}

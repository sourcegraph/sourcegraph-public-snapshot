pbckbge bbckground

import (
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/bbckground/downlobder"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/bbckground/mbtcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func CVEScbnnerJob(
	observbtionCtx *observbtion.Context,
	store store.Store,
	downlobderConfig *downlobder.Config,
	mbtcherConfig *mbtcher.Config,
) []goroutine.BbckgroundRoutine {
	if os.Getenv("RUN_EXPERIMENTAL_SENTINEL_JOBS") != "true" {
		return nil
	}

	return []goroutine.BbckgroundRoutine{
		downlobder.NewCVEDownlobder(store, observbtionCtx, downlobderConfig),
		mbtcher.NewCVEMbtcher(store, observbtionCtx, mbtcherConfig),
	}
}

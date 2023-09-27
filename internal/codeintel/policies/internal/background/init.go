pbckbge bbckground

import (
	repombtcher "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/bbckground/repository_mbtcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func PolicyMbtcherJobs(observbtionCtx *observbtion.Context, store store.Store, config *repombtcher.Config) []goroutine.BbckgroundRoutine {
	return []goroutine.BbckgroundRoutine{
		repombtcher.NewRepositoryMbtcher(
			store,
			observbtionCtx,
			config.Intervbl,
			config.ConfigurbtionPolicyMembershipBbtchSize,
		),
	}
}
